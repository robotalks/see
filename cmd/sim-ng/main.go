package main

import (
	"github.com/codingbrain/clix.go/exts/bind"
	"github.com/codingbrain/clix.go/exts/help"
	"github.com/codingbrain/clix.go/flag"
	"github.com/codingbrain/clix.go/term"
)

func main() {
	cli := &flag.CliDef{
		Cli: &flag.Command{
			Name: "sim-ng",
			Desc: "Simulation Engine",
			Options: []*flag.Option{
				&flag.Option{
					Name:    "port",
					Alias:   []string{"p"},
					Desc:    "Listening port",
					Tags:    map[string]interface{}{"help-var": "PORT"},
					Type:    "int",
					Default: 3500,
				},
				&flag.Option{
					Name:  "quiet",
					Alias: []string{"q"},
					Desc:  "Turn off the logs",
					Type:  "bool",
				},
				&flag.Option{
					Name: "version",
					Desc: "Show version and exit",
					Type: "bool",
				},
			},
			Commands: []*flag.Command{
				&flag.Command{
					Name:  "visualizer",
					Alias: []string{"vis"},
					Desc:  "Visualizer",
					Options: []*flag.Option{
						&flag.Option{
							Name:    "plugin-dir",
							Alias:   []string{"I"},
							Desc:    "Visualize plugin directory for object renders",
							Example: "-I plugin-dir1 -I plugin-dir2",
							List:    true,
							Tags:    map[string]interface{}{"help-var": "DIR"},
						},
						&flag.Option{
							Name: "web-content-dir",
							Desc: "Web content directory, for development",
							Type: "string",
						},
					},
					Arguments: []*flag.Option{
						&flag.Option{
							Name: "program",
							Desc: "Simulation program",
							Type: "string",
							Tags: map[string]interface{}{"help-var": "PROGRAM"},
						},
					},
				},
				&flag.Command{
					Name:  "version",
					Alias: []string{"ver"},
					Desc:  "Show Version",
				},
			},
		},
	}
	cli.Normalize()
	cli.Use(term.NewExt()).
		Use(bind.NewExt().
			Bind(&visCmd{}, "visualizer").
			Bind(&verCmd{}, "version")).
		Use(help.NewExt()).
		Parse().
		Exec()
}
