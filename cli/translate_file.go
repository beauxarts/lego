package cli

import (
	"bufio"
	"github.com/beauxarts/divido"
	"github.com/beauxarts/polyglot"
	"os"
	"time"
)

const (
	translatedExt = ".translated"
)

func translateFile(translator polyglot.Translator, filename, source, target string) error {

	resultFilename := translatedFilename(filename)
	if _, err := os.Stat(resultFilename); err == nil {
		return nil
	}

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
			if err.Error() == "429 Too Many Requests" {
				time.Sleep(time.Millisecond * 500)
			} else {
				return err
			}
		}

		tp.AddTranslatedContent(tc)
	}

	outf, err := os.Create(resultFilename)
	defer outf.Close()
	if err != nil {
		return err
	}

	if err := tp.Apply(outf); err != nil {
		return err
	}

	return nil
}

func translatedFilename(filename string) string {
	return filename + translatedExt
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
