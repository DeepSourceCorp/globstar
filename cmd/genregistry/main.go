package main

import (
	"fmt"
	"os"
	"context"
	"path/filepath"
	"github.com/urfave/cli/v3"
	
	"globstar.dev/checkers/discover"
)

func main() {
	app := &cli.Command {
		Name: "gen-registry",
		Usage: "Tool to dynamically generate checker registry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "dir",
				Aliases: []string{"d"},
				Usage: "Path to checker containing directory",
				Required: true,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			dirPath := cmd.String("dir")

			// verify if directory exists
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				return fmt.Errorf("directory '%s' does not exist", dirPath)
			}

			checkerDirs, err := getAllSubDirectories(dirPath)
			if err != nil {
				return fmt.Errorf("could not get all checker directories: %v", err)
			}

			err = discover.GenerateBuiltinChecker(checkerDirs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error generating registry: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}

	if err :=  app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "could not run gen-registry: %v", err)
	}
}

// helper function to get all subfolders
func getAllSubDirectories(topdir string) ([]string, error) {
	var subdirs []string
	entries, err := os.ReadDir(topdir)

	for _, entry := range entries {
		if entry.IsDir() {
			subdir := filepath.Join(topdir, entry.Name())
			subdirs = append(subdirs, subdir)
		}
	}

	return subdirs, err
}