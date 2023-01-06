package main

import (
	"bytes"
	_ "embed"
	"github.com/beauxarts/lego/cli"
	"github.com/boggydigital/clo"
	"github.com/boggydigital/nod"
	"log"
	"os"
)

var (
	//go:embed "cli-commands.txt"
	cliCommands []byte
	//go:embed "cli-help.txt"
	cliHelp []byte
)

func main() {

	nod.EnableStdOutPresenter()

	ns := nod.NewProgress("lego is creating audiobooks for your listening pleasure")
	defer ns.End()

	defs, err := clo.Load(
		bytes.NewBuffer(cliCommands),
		bytes.NewBuffer(cliHelp),
		nil)
	if err != nil {
		log.Fatalln(err)
	}

	clo.HandleFuncs(map[string]clo.Handler{
		"bind":       cli.BindHandler,
		"info":       cli.InfoHandler,
		"synthesize": cli.SynthesizeHandler,
	})

	if err := defs.AssertCommandsHaveHandlers(); err != nil {
		log.Fatalln(err)
	}

	if err := defs.Serve(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}
