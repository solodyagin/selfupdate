package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "selfupdatectl",
		Usage:       "A command line helper for various selfupdate tools.",
		Description: "The selfupdatectl command provides tooling for self updating Go applications.",
		Commands: []*cli.Command{
			createKeys(),
			sign(),
			check(),
			keyPrint(),
			awsUpload(),
			minioUpload(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
