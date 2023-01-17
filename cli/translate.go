package cli

import (
	"bufio"
	"github.com/beauxarts/polyglot"
	"github.com/beauxarts/polyglot/gcp"
	"github.com/boggydigital/nod"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func TranslateHandler(u *url.URL) error {
	q := u.Query()

	directory := q.Get("directory")
	source, target := q.Get("source"), q.Get("target")

	key := q.Get("key-value")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		}
	}

	return Translate(directory, source, target, key)
}

type xhtmlDecorations struct {
	prefix, suffix int
}

type translationPatch struct {
	sourceLines        []string
	contentDecorations map[int]xhtmlDecorations
	translatedContent  []string
}

func NewTranslationPatch(lines ...string) *translationPatch {
	return &translationPatch{
		sourceLines: lines,
	}
}

func (tp *translationPatch) UpdateContentDecorations() {

	tp.contentDecorations = make(map[int]xhtmlDecorations)

	for li, line := range tp.sourceLines {
		prefix, suffix := xhtmlDecorationsForLine(line)
		if prefix == -1 {
			continue
		}
		tp.contentDecorations[li] = xhtmlDecorations{prefix: prefix, suffix: suffix}
	}
}

func xhtmlDecorationsForLine(line string) (int, int) {
	prefix := strings.Index(line, ">")
	if prefix != -1 && prefix < len(line) {
		prefix += 1
	}
	suffix := strings.LastIndex(line, "<")
	if prefix > suffix {
		prefix = -1
	}
	return prefix, suffix
}

func (tp *translationPatch) SourceContent() []string {

	content := make([]string, 0, len(tp.contentDecorations))

	order := maps.Keys(tp.contentDecorations)
	sort.Ints(order)

	for _, li := range order {
		xd := tp.contentDecorations[li]
		if xd.prefix == -1 {
			continue
		}
		content = append(content, tp.sourceLines[li][xd.prefix:xd.suffix])
	}

	return content
}

func (tp *translationPatch) Apply(w io.Writer) error {

	index := 0
	for li, line := range tp.sourceLines {
		if xd, ok := tp.contentDecorations[li]; ok {
			// redecorate
			translatedLine := tp.translatedContent[index]

			line = line[:xd.prefix] + translatedLine + line[xd.suffix:]

			index++
		}

		io.WriteString(w, line+"\n")
	}

	return nil
}

const (
	epubXhtmlPattern = "/OEBPS/*.xhtml"
)

func Translate(directory, source, target, key string) error {

	ta := nod.NewProgress("translating epub files...")
	defer ta.End()

	files, err := filepath.Glob(filepath.Join(directory, epubXhtmlPattern))
	if err != nil {
		return ta.EndWithError(err)
	}

	ta.TotalInt(len(files))

	for _, filename := range files {
		if err := translateFile(filename, source, target, key); err != nil {
			return ta.EndWithError(err)
		}

		ta.Increment()
	}

	ta.EndWithResult("done")

	return nil
}

func translateFile(filename, source, target, key string) error {

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

	tp := NewTranslationPatch(lines...)

	tp.UpdateContentDecorations()

	contentLines := tp.SourceContent()

	if len(contentLines) == 0 {
		return nil
	}

	translator, err := gcp.NewTranslator(http.DefaultClient, gcp.NeuralMachineTranslation, key)
	if err != nil {
		return err
	}

	for from := 0; from < len(contentLines); from += 127 {

		to := minInt(from+127, len(contentLines))
		cl := contentLines[from:to]

		tc, err := translator.Translate(source, target, polyglot.HTML, cl...)
		if err != nil {
			return err
		}

		tp.translatedContent = append(tp.translatedContent, tc...)
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
