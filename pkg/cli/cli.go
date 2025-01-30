package cli

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/DeepSourceCorp/globstar/pkg/analysis"
	"github.com/DeepSourceCorp/globstar/pkg/config"
	"github.com/DeepSourceCorp/globstar/pkg/rules"
	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

type Cli struct {
	// RootDirectory is the target directory to analyze
	RootDirectory string
	// Rules is a list of lints that are applied to the files in `RootDirectory`
	Rules  []analysis.Rule
	Config *config.Config
}

func (c *Cli) loadConfig() error {
	conf, err := config.NewConfigFromFile(filepath.Join(c.RootDirectory, ".globstar", ".config.yml"))
	if err != nil {
		return err
	}

	c.Config = conf
	return nil
}

func (c *Cli) Run() error {
	err := c.loadConfig()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return err
	}

	// read all yaml rules in the .globstar directory
	patternRules, err := c.ReadCustomRules()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return err
	}

	cmd := &cli.Command{
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
						return c.RunLints(patternRules, false)
					} else if checkers == "builtin" {
						return c.RunLints(nil, true)
					} else if checkers == "all" || checkers == "" {
						return c.RunLints(patternRules, true)
					}
					return fmt.Errorf("Invalid value for --checkers flag, must be one of 'local', 'builtin' or 'all', got %s\n", checkers)
				},
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Run all tests in the `.globstar` directory",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					analysisDir := filepath.Join(c.RootDirectory, c.Config.RuleDir)
					passed, err := runTests(analysisDir)
					if err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
						return err
					}

					if !passed {
						return fmt.Errorf("tests failed")
					}

					fmt.Fprint(os.Stdout, "All tests passed")
					return nil
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

// ReadCustomRules reads all the custom rules from the `.globstar/` directory in the project root
func (c *Cli) ReadCustomRules() (map[analysis.Language][]analysis.YmlRule, error) {
	analysisDir := filepath.Join(c.RootDirectory, c.Config.RuleDir)

	stat, err := os.Stat(analysisDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	if !stat.IsDir() {
		return nil, nil
	}

	rulesMap := make(map[analysis.Language][]analysis.YmlRule)
	err = filepath.Walk(analysisDir, func(path string, d fs.FileInfo, err error) error {
		if d.IsDir() && d.Name() != c.Config.RuleDir {
			return fs.SkipDir
		}

		if d.Mode()&fs.ModeSymlink != 0 {
			// skip symlinks
			return nil
		}

		fileExt := filepath.Ext(path)
		isYamlFile := fileExt == ".yaml" || fileExt == ".yml"
		if !isYamlFile {
			return nil
		}

		patternRule, err := analysis.ReadFromFile(path)
		if err != nil {
			return fmt.Errorf("invalid rule '%s': %s", d.Name(), err.Error())
		}

		lang := patternRule.Language()
		rulesMap[lang] = append(rulesMap[lang], patternRule)
		return nil
	})

	return rulesMap, err
}

func (c *Cli) LintFile(
	rulesMap map[analysis.Language][]analysis.Rule,
	patternRules map[analysis.Language][]analysis.YmlRule,
	path string,
) ([]*analysis.Issue, error) {
	lang := analysis.LanguageFromFilePath(path)
	rules := rulesMap[lang]
	if rules == nil && patternRules == nil {
		// no rules are registered for this language
		return nil, nil
	}

	analyzer, err := analysis.FromFile(path, rules)
	if err != nil {
		return nil, err
	}
	analyzer.WorkDir = c.RootDirectory

	if patternRules != nil {
		analyzer.YmlRules = patternRules[lang]
	}

	return analyzer.Analyze(), nil
}

type lintResult struct {
	issues          []*analysis.Issue
	numFilesChecked int
}

func (lr *lintResult) GetExitStatus(conf *config.Config) int {
	for _, issue := range lr.issues {
		fmt.Println(issue.Category, issue.Severity)
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
	".globstar", // may contain test files
}

// RunLints goes over all the files in the project and runs the lints for every file encountered
func (c *Cli) RunLints(
	patternRules map[analysis.Language][]analysis.YmlRule, // map of language id -> yaml rules
	runBuiltinRules bool,
) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var allRules map[analysis.Language][]analysis.Rule
	if runBuiltinRules {
		allRules = rules.CreateBaseRuleMap()
	}

	result := lintResult{}
	err := filepath.Walk(c.RootDirectory, func(path string, d fs.FileInfo, err error) error {
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

		// run linter
		issues, err := c.LintFile(allRules, patternRules, path)
		if err != nil {
			// TODO: parse error on a single file should not exit the entire analysis process
			return err
		}

		if len(issues) > 0 {
			relPath, _ := filepath.Rel(c.RootDirectory, path)
			log.Error().Msgf("Issues found in %s", color.YellowString(relPath))
		}

		for _, issue := range issues {
			log.Error().Msgf("\t[Ln %d:Col %d] %s",
				issue.Range.StartPoint.Row,
				issue.Range.StartPoint.Column,
				color.YellowString(issue.Message),
			)

			result.issues = append(result.issues, issue)
		}

		return nil
	})

	if result.numFilesChecked > 0 {
		log.Info().Msgf("Analyzed %d files and found %d issues.", result.numFilesChecked, len(result.issues))
	} else {
		log.Info().Msg("No files to analyze")
	}

	exitStatus := result.GetExitStatus(c.Config)
	if exitStatus != 0 {
		fmt.Fprintf(os.Stderr, "Found %d issues\n", len(result.issues))
		return fmt.Errorf("Found %d issues", len(result.issues))
	}

	return err
}
