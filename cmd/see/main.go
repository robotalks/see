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
			Name: "see",
			Desc: "Visualization Engine",
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
					Name:    "plugin-dir",
					Alias:   []string{"I"},
					Desc:    "Visualize plugin directory for object renders",
					Example: "-I plugin-dir1 -I plugin-dir2",
					List:    true,
					Tags:    map[string]interface{}{"help-var": "DIR"},
				},
				&flag.Option{
					Name: "title",
					Desc: "Title for web page",
					Type: "string",
				},
				&flag.Option{
					Name: "version",
					Desc: "Show version and exit",
					Type: "bool",
				},
			},
			Arguments: []*flag.Option{
				&flag.Option{
					Name: "source",
					Desc: "Message source, can be a program or a URL\n" +
						"Supported protocol:\n" +
						"   MQHUB: mqhub://server:port/topic-prefix SCHEMA-FILE\n" +
						"   MQTT:  mqtt://server:port/topic-prefix\n",
					Type: "string",
					Tags: map[string]interface{}{"help-var": "SOURCE"},
				},
			},
		},
	}
	cli.Normalize()
	cli.Use(term.NewExt()).
		Use(bind.NewExt().Bind(&visCmd{})).
		Use(help.NewExt()).
		Parse().
		Exec()
}
