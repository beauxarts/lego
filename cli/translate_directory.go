package cli

import (
	"errors"
	"github.com/beauxarts/polyglot"
	"github.com/beauxarts/polyglot/acs"
	"github.com/beauxarts/polyglot/gcp"
	"github.com/boggydigital/nod"
	"net/http"
	"os"
	"path/filepath"
)

var epubHtmlPatterns = []string{
	"/OEBPS/*.xhtml",
	"/*.xhtml",
	"/OEBPS/*.html",
	"/*.html",
	"/html/*.html"}

func translateDirectory(directory, provider, from, to, key string) error {

	ta := nod.NewProgress("translating epub files...")
	defer ta.Done()

	var files []string

	for _, pattern := range epubHtmlPatterns {
		var err error
		files, err = filepath.Glob(filepath.Join(directory, pattern))
		if err != nil {
			return err
		}
		if len(files) > 0 {
			break
		}
	}

	ta.TotalInt(len(files))

	var translator polyglot.Translator
	var err error

	switch provider {
	case "gcp":
		translator, err = gcp.NewTranslator(http.DefaultClient, gcp.NeuralMachineTranslation, key)
	case "acs":
		translator, err = acs.NewTranslator(http.DefaultClient, key)
	default:
		err = errors.New("unknown provider " + provider)
	}

	if err != nil {
		return err
	}

	for _, filename := range files {
		if err := translateFile(translator, filename, from, to); err != nil {
			return err
		}
		ta.Increment()
	}

	// moving translated files over originals
	for _, filename := range files {
		resultFilename := translatedFilename(filename)
		if err := os.Rename(resultFilename, filename); err != nil {
			return err
		}
	}

	return nil
}
