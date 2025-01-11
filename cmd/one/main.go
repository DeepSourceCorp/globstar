package main

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/gobwas/glob"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/srijan-paul/deepgrep/pkg/one"
	"github.com/srijan-paul/deepgrep/pkg/rules"
	"github.com/urfave/cli/v3"
)

func LintFile(rulesMap map[one.Language][]one.Rule, path string) ([]*one.Issue, error) {
	rules := rulesMap[one.LanguageFromFilePath(path)]
	if rules == nil {
		// no rules are registered for this language
		return nil, nil
	}

	analyzer, err := one.FromFile(path, rules)
	if err != nil {
		return nil, err
	}

	return analyzer.Analyze(), nil
}

type lintResult struct {
	issues          []*one.Issue
	numFilesChecked int
}

func RunLints(rootDir string, globPattern string) error {
	ignorePattern, err := glob.Compile(globPattern)
	if err != nil {
		return err
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	allRules := rules.CreateRules()

	result := lintResult{}
	err = filepath.Walk(rootDir, func(path string, d fs.FileInfo, err error) error {
		if d.IsDir() {
			if ignorePattern.Match(path) {
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
		issues, err := LintFile(allRules, path)
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
					return RunLints(rootDir, ignorePattern)
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
