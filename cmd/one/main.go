package main

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

// ReadCustomRules reads all the custom rules from the `.one/` directory in the project root
func ReadCustomRules(projectRoot string) (map[one.Language][]one.PatternRule, error) {
	oneDir := filepath.Join(projectRoot, ".one")

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

	rulesMap := make(map[one.Language][]one.PatternRule)
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

func LintFile(
	rulesMap map[one.Language][]one.Rule,
	patternRules map[one.Language][]one.PatternRule,
	workDir string,
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
	analyzer.WorkDir = workDir

	if patternRules != nil {
		analyzer.PatternRules = patternRules[lang]
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
}

// RunLints goes over all the files in the project and runs the lints for every file encountered
func RunLints(
	rootDir string,
	globPattern string, // pattern to ignore files
	patternRules map[one.Language][]one.PatternRule, // map of language id -> yaml rules
) error {
	ignorePattern, err := glob.Compile(globPattern)
	if err != nil {
		return err
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	allRules := rules.CreateRules()

	result := lintResult{}
	err = filepath.Walk(rootDir, func(path string, d fs.FileInfo, err error) error {
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
		issues, err := LintFile(allRules, patternRules, rootDir, path)
		if err != nil {
			// TODO: parse error on a single file should not exit the entire analysis process
			return err
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

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to get current working directory")
		return
	}

	patternRules, err := ReadCustomRules(cwd)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return
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
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					rootDir, err := os.Getwd()
					if err != nil {
						return err
					}

					ignorePattern := cmd.String("ignore")
					return RunLints(rootDir, ignorePattern, patternRules)
				},
			},
			{
				Name:    "desc",
				Aliases: []string{"d"},
				Usage:   "Describe an issue",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Err(err)
	}
}
