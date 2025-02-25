package cli

import (
	"bytes"
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
	"globstar.dev/pkg/analysis"
	"globstar.dev/pkg/config"
)

type Cli struct {
	// RootDirectory is the target directory to analyze
	RootDirectory string
	// Rules is a list of lints that are applied to the files in `RootDirectory`
	Rules  []analysis.Rule
	Config *config.Config
	// DependencyGraph for the codebase
	DependencyGraph *analysis.DependencyGraph
}


func (c *Cli) loadConfig() error {
	conf, err := config.NewConfigFromFile(filepath.Join(c.RootDirectory, ".globstar", ".config.yml"))
	if err != nil {
		return err
	}

	c.Config = conf
	return nil
}

// GetChangedFiles returns the list of files that have been changed in the codebase
func (c *Cli) GetChangedFiles() ([]string, error) {

	cmd := exec.Command("git", "status", "--porcelain=v1", "-z")
	
	// Set working directory
	cmd.Dir = c.RootDirectory
	
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run git status in directory %s: %v", cmd.Dir, err)
	}

	var files []string
	
	// Parse the output
	for _, line := range strings.Split(strings.TrimRight(out.String(), "\x00"), "\x00") {
		if len(line) < 3 {
			continue
		}
		
		status := parseGitStatus(line[:2])
		path := strings.TrimSpace(line[3:])

		if status == "renamed" {
			parts := strings.Split(path, " -> ")
			if len(parts) == 2 {
				path = parts[1]
			}
		}

		if status == "untracked" || status == "deleted" || status == "unknown" {
			continue
		}
		
		if path != "" {
			// Convert path to absolute path
			absPath, err := filepath.Abs(filepath.Join(cmd.Dir, path))
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path for %s: %v", path, err)
			}
			
			files = append(files, absPath)
		}
	}

	return files, nil
}
// func (c *Cli) getModuleName() (string, error) {
// 	goModPath := filepath.Join(c.RootDirectory, "go.mod") 

// 	if _,err := os.Stat(goModPath); os.IsNotExist(err){
// 		return "", fmt.Errorf("Mod file not found")
// 	}


// 	content, err := os.ReadFile(goModPath)
// 	if err!= nil{
// 		return "", err
// 	}

// 	lines := strings.Split(string(content), "\n")
// 	for _, line := range lines{
// 		line = strings.TrimSpace(line)
// 		if strings.HasPrefix(line, "module ") {
// 			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
// 		}
// 	}

// 	return "", fmt.Errorf("module name oculd not be found")
// }

func (c *Cli) BuildDependencyGraph() error {
    c.DependencyGraph = analysis.NewDependencyGraph()
    
    // Walk through all files in the project to build the dependency graph
    err := filepath.Walk(c.RootDirectory, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil // continue to the next file
        }

        if info.IsDir() {
            if c.Config.ShouldExcludePath(path) {
                return filepath.SkipDir
            }
            return nil
        }

        // Skip excluded paths
        if c.Config.ShouldExcludePath(path) {
            return nil
        }

        // Only process files with known languages
        language := analysis.LanguageFromFilePath(path)
        if language == analysis.LangUnknown {
            return nil
        }

        // Extract dependencies for this file
        deps, err := analysis.ExtractDependencies(path)
        if err != nil {
            // Log but continue
            log.Debug().Err(err).Str("path", path).Msg("Failed to extract dependencies")
            return nil
        }

        // Add dependencies to the graph
        for _, dep := range deps {
            c.DependencyGraph.AddDependency(path, dep)
        }

        return nil
    })

    return err
}



func (c *Cli) GetAffectedFiles() ([]string, error) {
    changedFiles, err := c.GetChangedFiles()
    if err != nil {
        return nil, err
    }

    if c.DependencyGraph == nil {
        if err := c.BuildDependencyGraph(); err != nil {
            return nil, fmt.Errorf("failed to build dependency graph: %w", err)
        }
    }

    // Get all affected files
    affectedFiles := c.DependencyGraph.GetAffectedFiles(changedFiles)

    return affectedFiles, nil
}


func (c *Cli) Run() error {
	err := c.loadConfig()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return err
	}

	// read all yaml rules in the .globstar directory
	patternRules, err := c.LoadYamlRules(c.Config.RuleDir)
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
						return c.RunLints(patternRules, false)
					} else if checkers == "builtin" {
						return c.RunLints(nil, true)
					} else if checkers == "all" || checkers == "" {
						return c.RunLints(patternRules, true)
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
						dir = c.Config.RuleDir
					}
					analysisDir := filepath.Join(c.RootDirectory, dir)
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

// LoadYamlRules reads all the custom rules from the `.globstar/` directory in the project root,
// or from checkers/ dir for built-in rules
func (c *Cli) LoadYamlRules(ruleDir string) (map[analysis.Language][]analysis.YmlRule, error) {
	analysisDir := filepath.Join(c.RootDirectory, ruleDir)

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
		if err != nil {
			return nil
		}

		if d.IsDir() && d.Name() != ruleDir {
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

	if patternRules == nil {
		patternRules = make(map[analysis.Language][]analysis.YmlRule)
	}

	var goAnalyzers []*goAnalysis.Analyzer
	if runBuiltinRules {
		goAnalyzers = checkers.LoadGoRules()
		builtInPatternRules, err := checkers.LoadYamlRules()
		if err != nil {
			return err
		}

		// merge the built-in rules with the custom rules
		for lang, rules := range builtInPatternRules {
			if _, ok := patternRules[lang]; ok {
				patternRules[lang] = append(patternRules[lang], rules...)
			} else {
				patternRules[lang] = rules
			}
		}
	}
	err := c.BuildDependencyGraph()
	// c.DependencyGraph.PrintGraph()
	if err != nil{
		return fmt.Errorf("could not buuild dep. graph: %v", err)
	}

	// git integration to track changed files.
	changedFiles, err := c.GetChangedFiles()
	if err != nil {
		return fmt.Errorf("Error building dependency graph: %v\n", err)
	}

	for _, file := range changedFiles {
		err := c.DependencyGraph.GetFileDependencies(file)	
		if err != nil {
			return err
		}
	}	



	result := lintResult{}
	err = filepath.Walk(c.RootDirectory, func(path string, d fs.FileInfo, err error) error {
		if err != nil {
			// skip this path
			return nil
		}

		if !slices.Contains(changedFiles, path) {
			// skip this path(file), if it's not changed/modified
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

		// run linter
		// the first arg is empty, since the format for inbuilt Go-based rules has changed
		// TODO: factor it in later
		issues, err := c.LintFile(map[analysis.Language][]analysis.Rule{}, patternRules, path)
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

	goIssues, err := goAnalysis.RunAnalyzers(c.RootDirectory, goAnalyzers)
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

	// FIXME: go based rules do not increment the numFilesChecked counter
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


func parseGitStatus(code string) string {
	code = strings.TrimSpace(code)
	switch code {
	case "M", "MM":
		return "modified"
	case "A":
		return "added"
	case "D":
		return "deleted"
	case "R":
		return "renamed"
	case "??":
		return "untracked"
	default:
		return "unknown"
	}
}