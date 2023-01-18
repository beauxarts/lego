package cli

import (
	"bufio"
	"errors"
	"github.com/beauxarts/divido"
	"github.com/beauxarts/polyglot"
	"github.com/beauxarts/polyglot/acs"
	"github.com/beauxarts/polyglot/gcp"
	"github.com/boggydigital/nod"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func TranslateHandler(u *url.URL) error {
	q := u.Query()

	directory := q.Get("directory")
	from, to := q.Get("from"), q.Get("to")

	key := q.Get("key-value")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		}
	}

	provider := q.Get("provider")

	return Translate(directory, provider, from, to, key)
}

var epubHtmlPatterns = []string{"/OEBPS/*.xhtml", "/*.xhtml", "/html/*.html"}

func Translate(directory, provider, from, to, key string) error {

	ta := nod.NewProgress("translating epub files...")
	defer ta.End()

	var files []string

	for _, pattern := range epubHtmlPatterns {
		var err error
		files, err = filepath.Glob(filepath.Join(directory, pattern))
		if err != nil {
			return ta.EndWithError(err)
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
		return ta.EndWithError(err)
	}

	for _, filename := range files {
		if err := translateFile(translator, filename, from, to, key); err != nil {
			return ta.EndWithError(err)
		}

		ta.Increment()
	}

	ta.EndWithResult("done")

	return nil
}

func translateFile(translator polyglot.Translator, filename, source, target, key string) error {

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	lines := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := file.Close(); err != nil {
		return err
	}

	tp := divido.NewTranslationPatch(lines...)

	tp.UpdateContentDecorations()

	contentLines := tp.SourceContent()

	if len(contentLines) == 0 {
		return nil
	}

	// break into chunks to account for 128 strings limit
	for from := 0; from < len(contentLines); from += 127 {

		to := minInt(from+127, len(contentLines))
		cl := contentLines[from:to]

		format := polyglot.Text
		if translator.IsHTMLSupported() {
			format = polyglot.HTML
		}

		tc, err := translator.Translate(source, target, format, cl...)
		if err != nil {
			return err
		}

		tp.AddTranslatedContent(tc)
	}

	outf, err := os.Create(filename)
	defer outf.Close()
	if err != nil {
		return err
	}

	if err := tp.Apply(outf); err != nil {
		return err
	}

	return nil
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
