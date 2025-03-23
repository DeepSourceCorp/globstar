package cli

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	goAnalysis "globstar.dev/analysis"
	"globstar.dev/checkers"
	"globstar.dev/checkers/discover"
	"globstar.dev/pkg/analysis"
	"globstar.dev/pkg/config"
	"globstar.dev/util"
)

type Cli struct {
	// RootDirectory is the target directory to analyze
	RootDirectory string
	// Checkers is a list of checkers that are applied to the files in `RootDirectory`
	Checkers []analysis.Checker
	Config   *config.Config
}

func (c *Cli) loadConfig() error {
	conf, err := config.NewConfigFromFile(filepath.Join(c.RootDirectory, ".globstar", ".config.yml"))
	if err != nil {
		return err
	}

	c.Config = conf
	return nil
}

func (c *Cli) runCustomGoAnalyzerTests() (bool, error) {
	if err := c.buildCustomGoCheckers(); err != nil {
		return false, err
	}

	if _, err := os.Stat(filepath.Join(c.RootDirectory, "custom-analyzer")); err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}

		return false, err
	}

	_, stderr, err := util.RunCmd("./custom-analyzer", []string{"-test", "-path", filepath.Join(c.RootDirectory, c.Config.CheckerDir)}, c.RootDirectory)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			fmt.Fprintln(os.Stderr, stderr)
			return false, nil
		}
		fmt.Fprintf(os.Stderr, "Error running custom Go-based tests: %s\n", err)
		return false, err
	}

	fmt.Fprintln(os.Stderr, stderr)
	return true, nil
}

func (c *Cli) runCustomGoAnalyzers() ([]*goAnalysis.Issue, []string, error) {

	issues := []*goAnalysis.Issue{}
	issuesAsText := []string{}

	if err := c.buildCustomGoCheckers(); err != nil {
		return issues, issuesAsText, err
	}

	if _, err := os.Stat(filepath.Join(c.RootDirectory, "custom-analyzer")); err != nil {
		if os.IsNotExist(err) {
			return issues, issuesAsText, nil
		}

		return issues, issuesAsText, err
	}

	_, stderr, err := util.RunCmd("./custom-analyzer", []string{"-path", filepath.Join(c.RootDirectory, c.Config.CheckerDir)}, c.RootDirectory)
	if err != nil && err.(*exec.ExitError).ExitCode() != 1 {
		return issues, issuesAsText, err
	}

	scanner := bufio.NewScanner(strings.NewReader(stderr))
	for scanner.Scan() {
		scannedIssue := []byte(scanner.Text())
		issue, err := goAnalysis.IssueFromJson(scannedIssue)
		if err != nil {
			continue
		}
		issues = append(issues, issue)

		txt, _ := goAnalysis.IssueAsTextFromJson(scannedIssue)
		issuesAsText = append(issuesAsText, string(txt))
	}

	return issues, issuesAsText, nil
}

func (c *Cli) Run() error {
	err := c.loadConfig()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return err
	}

	cli.VersionPrinter = func(cmd *cli.Command) {
		version := strings.TrimPrefix(cmd.Version, "v")
		fmt.Println(version)
	}

	cmd := &cli.Command{
		Name:    "globstar",
		Usage:   "The open-source static analysis toolkit",
		Version: version,
		Description: `Globstar helps you write and run custom checkers for bad and insecure patterns and run them on
your codebase with a simple command. It comes with built-in checkers that you can use out-of-the-box,\
or you can write your own in the .globstar directory of any repository.`,
		Commands: []*cli.Command{
			{
				Name:    "check",
				Aliases: []string{"c"},
				Usage:   "Run Globstar on the current project",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "ignore",
						Usage:   "Ignore file paths that match a pattern",
						Aliases: []string{"i"},
					},

					&cli.StringFlag{
						Name: "checkers",
						Usage: `Specify whether to run the built-in checkers, the local checkers
(in the .globstar directory) or both. Use --checkers=local to run only the local checkers, --checkers=builtin
to run only the built-in checkers, and --checkers=all to run both.`,
						Aliases: []string{"c"},
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					ignorePattern := cmd.String("ignore")
					if err := c.Config.AddExcludePatterns(ignorePattern); err != nil {
						return err
					}

					checkers := cmd.String("checkers")
					if checkers == "local" {
						return c.RunCheckers(false, true)
					} else if checkers == "builtin" {
						return c.RunCheckers(true, false)
					} else if checkers == "all" || checkers == "" {
						return c.RunCheckers(true, true)
					}
					return fmt.Errorf("invalid value for --checkers flag, must be one of 'local', 'builtin' or 'all', got %s", checkers)
				},
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Run all tests in the specified directory. If no directory is specified, the tests are run in the `.globstar` directory.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "directory",
						Usage:   "Specify the directory to run tests in",
						Aliases: []string{"d"},
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dir := cmd.String("directory")
					if dir == "" {
						dir = c.Config.CheckerDir
					}
					analysisDir := filepath.Join(c.RootDirectory, dir)

					// Track test failures but continue running all tests
					var testsFailed bool

					yamlPassed, err := runTests(analysisDir)
					if err != nil {
						err = fmt.Errorf("error running YAML tests: %w", err)
						fmt.Fprintln(os.Stderr, err.Error())
						// Don't return immediately, continue with other tests
					}
					if !yamlPassed {
						testsFailed = true
					}

					goPassed := true
					if dir == "checkers" {
						var errs []error
						goPassed, errs = checkers.RunAnalyzerTests(checkers.AnalyzerRegistry)
						if len(errs) > 0 {
							fmt.Fprintln(os.Stderr, "Errors running Go-based tests:")
							for _, e := range errs {
								fmt.Fprintln(os.Stderr, e.Error())
								testsFailed = true
							}
						}
					} else {
						goPassed, err = c.runCustomGoAnalyzerTests()
						if err != nil {
							fmt.Fprintln(os.Stderr, err.Error())
							testsFailed = true
						}
					}
					if !goPassed {
						testsFailed = true
					}

					if testsFailed {
						return fmt.Errorf("one or more tests failed")
					}

					fmt.Fprint(os.Stdout, "All tests passed")
					return nil
				},
			},
			{
				Name:    "build",
				Aliases: []string{"b"},
				Usage:   "Build the custom Go checkers in the .globstar directory",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return c.buildCustomGoCheckers()
				},
			},
		},
	}

	err = cmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}

	return err
}

func (c *Cli) buildCustomGoCheckers() error {
	// verify that the checker directory exists
	if _, err := os.Stat(c.Config.CheckerDir); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Checker directory %s does not exist\n", c.Config.CheckerDir)
			return nil
		}
		return nil
	}

	if goCheckers, err := discover.DiscoverGoCheckers(c.Config.CheckerDir); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	} else {
		if len(goCheckers) == 0 {
			fmt.Fprintln(os.Stderr, "No Go checkers found in the directory")
			return nil
		}
	}

	tempDir, err := os.MkdirTemp("", "build")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	defer os.RemoveAll(tempDir)

	err = discover.GenerateAnalyzer(c.Config.CheckerDir, tempDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	err = discover.BuildAnalyzer(tempDir, c.RootDirectory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	return nil
}

func (c *Cli) CheckFile(
	checkersMap map[analysis.Language][]analysis.Checker,
	patternCheckers map[analysis.Language][]analysis.YamlChecker,
	path string,
) ([]*analysis.Issue, error) {
	lang := analysis.LanguageFromFilePath(path)
	checkers := checkersMap[lang]
	if checkers == nil && patternCheckers == nil {
		// no checkers are registered for this language
		return nil, nil
	}

	analyzer, err := analysis.FromFile(path, checkers)
	if err != nil {
		return nil, err
	}
	analyzer.WorkDir = c.RootDirectory

	if patternCheckers != nil {
		analyzer.YamlCheckers = patternCheckers[lang]
	}

	return analyzer.Analyze(), nil
}

type checkResult struct {
	issues          []*analysis.Issue
	numFilesChecked int
}

func (lr *checkResult) GetExitStatus(conf *config.Config) int {
	for _, issue := range lr.issues {
		for _, failCategory := range conf.FailWhen.CategoryIn {
			if issue.Category == failCategory {
				return conf.FailWhen.ExitCode
			}
		}

		for _, failSeverity := range conf.FailWhen.SeverityIn {
			if issue.Severity == failSeverity {
				return conf.FailWhen.ExitCode
			}
		}
	}

	return 0
}

var defaultIgnoreDirs = []string{
	"checkers",
	"node_modules",
	"vendor",
	"dist",
	"build",
	"out",
	".git",
	".svn",
	"venv",
	"__pycache__",
	".idea",
}

// RunCheckers goes over all the files in the project and runs the checks for every file encountered
func (c *Cli) RunCheckers(runBuiltinCheckers, runCustomCheckers bool) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	patternCheckers := make(map[analysis.Language][]analysis.YamlChecker)

	var goAnalyzers []*goAnalysis.Analyzer
	if runBuiltinCheckers {
		goAnalyzers = checkers.LoadGoCheckers()
		builtInPatternCheckers, err := checkers.LoadBuiltinYamlCheckers()
		if err != nil {
			return err
		}

		// merge the built-in checkers with the custom checkers
		for lang, checkers := range builtInPatternCheckers {
			if _, ok := patternCheckers[lang]; ok {
				patternCheckers[lang] = append(patternCheckers[lang], checkers...)
			} else {
				patternCheckers[lang] = checkers
			}
		}
	}

	if runCustomCheckers {
		customYamlCheckers, err := checkers.LoadCustomYamlCheckers(c.Config.CheckerDir)
		if err != nil {
			return err
		}

		// merge customYamlCheckers into yamlCheckers
		for lang, checkers := range customYamlCheckers {
			if _, ok := patternCheckers[lang]; ok {
				patternCheckers[lang] = append(patternCheckers[lang], checkers...)
			} else {
				patternCheckers[lang] = checkers
			}
		}
	}

	result := checkResult{}
	err := filepath.Walk(c.RootDirectory, func(path string, d fs.FileInfo, err error) error {
		if err != nil {
			// skip this path
			return nil
		}

		if d.IsDir() {
			if c.Config.ShouldExcludePath(path) || slices.Contains(defaultIgnoreDirs, d.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if d.Mode()&fs.ModeSymlink != 0 {
			// skip symlinks
			return nil
		}

		if c.Config.ShouldExcludePath(path) {
			return nil
		}

		language := analysis.LanguageFromFilePath(path)
		if language == analysis.LangUnknown {
			return nil
		}

		result.numFilesChecked++

		// run checker
		// the first arg is empty, since the format for inbuilt Go-based checkers has changed
		// TODO: factor it in later
		issues, err := c.CheckFile(map[analysis.Language][]analysis.Checker{}, patternCheckers, path)
		if err != nil {
			// parse error on a single file should not exit the entire analysis process
			// TODO: logging the below error message is not helpful, as it logs unsupported file types as well
			// fmt.Fprintf(os.Stderr, "Error parsing file %s: %s\n", path, err)
			return nil
		}

		for _, issue := range issues {
			txt, _ := issue.AsText()
			log.Error().Msg(string(txt))

			result.issues = append(result.issues, issue)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(goAnalyzers) > 0 {
		goIssues, err := goAnalysis.RunAnalyzers(c.RootDirectory, goAnalyzers, nil)
		if err != nil {
			return fmt.Errorf("failed to run Go-based analyzers: %w", err)
		}
		for _, issue := range goIssues {
			txt, _ := issue.AsText()
			log.Error().Msg(string(txt))

			result.issues = append(result.issues, &analysis.Issue{
				Filepath: issue.Filepath,
				Message:  issue.Message,
				Severity: config.Severity(issue.Severity),
				Category: config.Category(issue.Category),
				Node:     issue.Node,
				Id:       issue.Id,
			})
		}
	}

	if runCustomCheckers {
		customGoIssues, textIssues, err := c.runCustomGoAnalyzers()
		if err != nil {
			return fmt.Errorf("failed to run custom Go-based analyzers: %w", err)
		}

		for _, txt := range textIssues {
			log.Error().Msg(string(txt))
		}

		for _, issue := range customGoIssues {
			result.issues = append(result.issues, &analysis.Issue{
				Filepath: issue.Filepath,
				Message:  issue.Message,
				Severity: config.Severity(issue.Severity),
				Category: config.Category(issue.Category),
				Node:     issue.Node,
				Id:       issue.Id,
			})
		}
	}

	// FIXME: go based checkers do not increment the numFilesChecked counter
	if result.numFilesChecked > 0 {
		log.Info().Msgf("Analyzed %d files and found %d issues.", result.numFilesChecked, len(result.issues))
	} else {
		log.Info().Msg("No files to analyze")
	}

	exitStatus := result.GetExitStatus(c.Config)
	if exitStatus != 0 {
		fmt.Fprintf(os.Stderr, "Found %d issues\n", len(result.issues))
		return fmt.Errorf("found %d issues", len(result.issues))
	}

	return err
}
