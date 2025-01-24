package cli

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/fatih/color"
	"github.com/gobwas/glob"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/srijan-paul/deepgrep/pkg/one"
	"github.com/srijan-paul/deepgrep/pkg/rules"
	"github.com/urfave/cli/v3"
)

type Cli struct {
	// RootDirectory is the target directory to analyze
	RootDirectory string
	// Rules is a list of lints that are applied to the files in `RootDirectory`
	Rules []one.Rule
}

func (c *Cli) Run() error {
	// read all yaml rules in the .one directory
	patternRules, err := c.ReadCustomRules()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return err
	}

	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "lint",
				Aliases: []string{"l"},
				Usage:   "Run OneLint on the current project",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "ignore",
						Usage:   "Ignore file paths that match a pattern",
						Aliases: []string{"i"},
					},

					&cli.BoolFlag{
						Name:    "builtins",
						Usage:   "Run all the builtin rules",
						Aliases: []string{"b"},
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					ignorePattern := cmd.String("ignore")
					useBuiltins := cmd.Bool("builtins")

					return c.RunLints(ignorePattern, patternRules, useBuiltins)
				},
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Run all tests in the `.one` directory",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					oneDir := filepath.Join(c.RootDirectory, ".one")
					passed, err := runTests(oneDir)
					if err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
						return err
					}

					if !passed {
						return fmt.Errorf("tests failed")
					}

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

// ReadCustomRules reads all the custom rules from the `.one/` directory in the project root
func (c *Cli) ReadCustomRules() (map[one.Language][]one.YmlRule, error) {
	oneDir := filepath.Join(c.RootDirectory, ".one")

	stat, err := os.Stat(oneDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	if !stat.IsDir() {
		return nil, nil
	}

	rulesMap := make(map[one.Language][]one.YmlRule)
	err = filepath.Walk(oneDir, func(path string, d fs.FileInfo, err error) error {
		if d.IsDir() && d.Name() != ".one" {
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

		patternRule, err := one.ReadFromFile(path)
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
	rulesMap map[one.Language][]one.Rule,
	patternRules map[one.Language][]one.YmlRule,
	path string,
) ([]*one.Issue, error) {
	lang := one.LanguageFromFilePath(path)
	rules := rulesMap[lang]
	if rules == nil && patternRules == nil {
		// no rules are registered for this language
		return nil, nil
	}

	analyzer, err := one.FromFile(path, rules)
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
	issues          []*one.Issue
	numFilesChecked int
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
	".one", // may contain test files
}

// RunLints goes over all the files in the project and runs the lints for every file encountered
func (c *Cli) RunLints(
	globPattern string, // pattern to ignore files
	patternRules map[one.Language][]one.YmlRule, // map of language id -> yaml rules
	runBuiltinRules bool,
) error {
	ignorePattern, err := glob.Compile(globPattern)
	if err != nil {
		return err
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var allRules map[one.Language][]one.Rule
	if runBuiltinRules {
		allRules = rules.CreateBaseRuleMap()
	}

	result := lintResult{}
	err = filepath.Walk(c.RootDirectory, func(path string, d fs.FileInfo, err error) error {
		if d.IsDir() {
			if ignorePattern.Match(path) || slices.Contains(defaultIgnoreDirs, d.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if d.Mode()&fs.ModeSymlink != 0 {
			// skip symlinks
			return nil
		}

		if ignorePattern.Match(path) {
			return nil
		}

		language := one.LanguageFromFilePath(path)
		if language == one.LangUnknown {
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
			log.Error().Msgf("[Ln %d:Col %d] %s",
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

	return err
}
