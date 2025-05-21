package main

import (
	"bytes"
	_ "embed"
	"github.com/beauxarts/binder/cli"
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

	ns := nod.NewProgress("binder is packing your audiobooks")
	defer ns.Done()

	defs, err := clo.Load(
		bytes.NewBuffer(cliCommands),
		bytes.NewBuffer(cliHelp),
		nil)
	if err != nil {
		log.Fatalln(err)
	}

	clo.HandleFuncs(map[string]clo.Handler{
		"bind-book":                 cli.BindBookHandler,
		"bind-chapters":             cli.BindChaptersHandler,
		"chapter-metadata":          cli.ChapterMetadataHandler,
		"cover":                     cli.CoverHandler,
		"pack-audiobook":            cli.PackAudiobookHandler,
		"prepare-external-chapters": cli.PrepareExternalChaptersHandler,
	})

	if err = defs.AssertCommandsHaveHandlers(); err != nil {
		log.Fatalln(err)
	}

	if err = defs.Serve(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}
