package cli

import (
	"github.com/beauxarts/divido"
	"github.com/beauxarts/lego/chapter_paragraph"
	"github.com/boggydigital/nod"
	"golang.org/x/exp/slices"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var breakAliases = []string{
	"* * *",
}

func SynthesizeHandler(u *url.URL) error {

	q := u.Query()

	textFilename := q.Get("text-filename")
	outputDirectory := q.Get("output-directory")

	provider := q.Get("provider")
	region := q.Get("region")

	key := q.Get("key-value")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		}
	}

	voiceParams := strings.Split(q.Get("voice-params"), ",")

	overwrite := q.Has("overwrite")

	return Synthesize(textFilename, outputDirectory, provider, region, key, voiceParams, overwrite)
}

func Synthesize(
	textFilename string,
	outputDirectory string,
	provider, region, key string,
	voiceParams []string,
	overwrite bool) error {
	sa := nod.NewProgress("synthesizing chapter paragraphs from text...")
	defer sa.End()

	//in order to convert text file to audiobook the following steps are required:
	//- process text document to identify chapters, paragraphs
	//- synthesize chapter title named 00000000c-000000000.ogg
	//- synthesize chapter by chapter, paragraph by paragraph to create files named 00000000c-00000000p.ogg
	//- create a list of chapter paragraph audio files 00000000c.txt

	file, err := os.Open(textFilename)
	defer file.Close()
	if err != nil {
		return sa.EndWithError(err)
	}

	var td divido.TextDocument

	notesFilename := divido.DefaultNotesFilename(textFilename)
	if _, err := os.Stat(notesFilename); err == nil {
		notes, err := os.Open(notesFilename)
		defer notes.Close()
		if err != nil {
			return sa.EndWithError(err)
		}
		td = divido.NewTextDocumentWithNotes(file, notes)
	} else {
		td = divido.NewTextDocument(file)
	}

	chapters := td.ChapterTitles()

	var szr *chapter_paragraph.Synthesizer

	switch provider {
	case "acs":
		szr, err = chapter_paragraph.NewACSSynthesizer(http.DefaultClient, voiceParams, region, key, outputDirectory, overwrite)
	case "gcp":
		szr, err = chapter_paragraph.NewGCPSynthesizer(http.DefaultClient, voiceParams, key, outputDirectory, overwrite)
	case "say":
		szr, err = chapter_paragraph.NewSaySynthesizer(voiceParams, outputDirectory, overwrite)
	}

	if err != nil {
		return sa.EndWithError(err)
	}

	sa.TotalInt(len(chapters))

	for ci, ct := range chapters {

		if err := szr.CreateChapterTitle(ci, ct); err != nil {
			return sa.EndWithError(err)
		}

		pa := nod.NewProgress(" synthesizing chapter %d paragraphs...", ci+1)

		paragraphs := td.ChapterParagraphs(ct)
		pa.TotalInt(len(paragraphs))

		for pi, pt := range paragraphs {

			pts := strings.TrimSpace(string(pt))
			if slices.Contains(breakAliases, pts) {
				if err = szr.CreatePause(ci, pi); err != nil {
					return pa.EndWithError(err)
				}
				continue
			}

			if err = szr.CreateChapterParagraph(ci, pi, pts); err != nil {
				return pa.EndWithError(err)
			}
			pa.Increment()
		}

		pa.EndWithResult("done")

		if err = szr.CreateChapterFilesList(ci, len(paragraphs)); err != nil {
			return sa.EndWithError(err)
		}

		sa.Increment()
	}

	if err := szr.CreateChapters(td.ChapterTitles()); err != nil {
		return sa.EndWithError(err)
	}

	sa.EndWithResult("done")

	return nil
}
