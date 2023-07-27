package main

import (
	"errors"
	"log"
	"os"

	"github.com/actiontech/dms/internal/dms/cmd/gencli/gen"

	"github.com/urfave/cli/v2"
)

// main is the main function
func main() {
	var debug bool

	app := &cli.App{
		Name:                 "gencli",
		HelpName:             "gencli for actiontech dms",
		Version:              "2.1.2",
		Usage:                "used for actiontech dms model generation",
		Description:          "cli for generating and executing migrations with gorm",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name: "generate-node-repo-fields",
				Aliases: []string{
					"gnrf",
				},
				ArgsUsage: "directory_to_search directory_write_to",
				Usage:     "to generate repo fields for nodes",
				Action: func(c *cli.Context) error {
					searchDirectory := c.Args().Get(0)
					writeDirectory := c.Args().Get(1)

					if searchDirectory == "" || writeDirectory == "" {
						return errors.New("must specify search directory and write directory")
					}

					if debug {
						log.Printf("generating repo fields from directory [%s]", searchDirectory)
					}

					return gen.GenRepoFieldsFile(debug, searchDirectory, writeDirectory)
				},
			},
		},
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "execute in debug mode",
				Value:       false,
				Destination: &debug,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
