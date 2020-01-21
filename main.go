package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/yuichiro12/envelope/cmd"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:      "list",
				Usage:     "list all parameters in aws parameter store with given prefix",
				UsageText: "envelope list /Myservice/MyApp/Dev",
				Action:    cmd.List,
			},
			{
				Name:      "apply",
				Usage:     "apply .env to aws parameter store with given prefix and filepath",
				UsageText: "envelope apply -f /path/to/.env /Myservice/MyApp/Dev",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Required: true,
					},
					&cli.BoolFlag{
						Name: "no-interactive",
						Aliases:  []string{"y"},
					},
				},
				Action: cmd.Apply,
			},
			{
				Name:      "diff",
				Usage:     "show diff before applying .env",
				UsageText: "envelope diff -f /path/to/.env /Myservice/MyApp/Dev",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Required: true,
					},
				},
				Action: cmd.Diff,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
