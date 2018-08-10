package main

import (
	"os"

	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	cli.AppHelpTemplate = helpTemplate
	app := cli.NewApp()
	app.Name = "gostub"
	app.Usage = "generate stubs for your Go interfaces"
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "source, s",
			Usage: "the source directory where the interface will be searched. If not specified, the current directory is used.",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "the output filepath where the stub will be saved. The name of the last directory in the filepath will determine the generated stub's package.",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "the name of the generated stub. If not specified, the 'Stub' suffix is appended to the interface name in order to form the stub name.",
		},
	}
	app.Action = RunGoStub
	app.Run(os.Args)
}

const helpTemplate = `
NAME
   {{.Name}} - {{.Usage}}

SYNOPSIS
   {{.Name}} [-s source_folder] [-o output_file] [-n stub_name] interface_name

DESCRIPTION
   The gostub command generates a Stub structure which implements the specified interface.

   {{range .Flags}}{{.}}
   {{end}}
`
