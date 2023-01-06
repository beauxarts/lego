package cli

import (
	"errors"
	"github.com/beauxarts/divido"
	gti "github.com/beauxarts/google_tts_integration"
	"github.com/beauxarts/lego/chapter_paragraph"
	"github.com/boggydigital/nod"
	"net/http"
	"net/url"
	"os"
)

func SynthesizeHandler(u *url.URL) error {

	q := u.Query()

	filename := q.Get("input-filename")
	key := q.Get("key")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		} else {
			return errors.New("key file not found")
		}
	}

	if key == "" {
		return errors.New("synthesis requires a key (or a key file)")
	}

	vl, vn, vg := q.Get("voice-locale"), q.Get("voice-name"), q.Get("voice-gender")
	voice := gti.NewVoice(vl, vn, vg)

	outputDirectory := q.Get("output-directory")
	overwrite := q.Has("overwrite")

	return Synthesize(filename, voice, key, outputDirectory, overwrite)
}

func Synthesize(inputFilename string, voice *gti.VoiceSelectionParams, key, outputDirectory string, overwrite bool) error {
	sa := nod.NewProgress("synthesizing audiobook from text...")
	defer sa.End()

	//in order to convert text file to audiobook the following steps are required:
	//- process text document to identify chapters, paragraphs
	//- synthesize chapter title named 00000000c-000000000.ogg
	//- synthesize chapter by chapter, paragraph by paragraph to create files named 00000000c-00000000p.ogg
	//- create a list of chapter paragraph audio files 00000000c.txt

	file, err := os.Open(inputFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	td := divido.NewTextDocument(file)
	chapters := td.ChapterTitles()

	cps := chapter_paragraph.NewSynthesizer(http.DefaultClient, voice, key, outputDirectory, overwrite)

	sa.TotalInt(len(chapters))

	for ci, ct := range chapters {

		if err := cps.CreateChapterTitle(ci, ct); err != nil {
			return sa.EndWithError(err)
		}

		pa := nod.NewProgress(" synthesizing chapter %d paragraphs...", ci+1)

		paragraphs := td.ChapterParagraphs(ct)
		pa.TotalInt(len(paragraphs))

		for pi, pt := range paragraphs {
			if err = cps.CreateChapterParagraph(ci, pi, string(pt)); err != nil {
				return pa.EndWithError(err)
			}
			pa.Increment()
		}

		pa.EndWithResult("done")

		if err = cps.CreateChapterFilesList(ci, len(paragraphs)); err != nil {
			return sa.EndWithError(err)
		}

		sa.Increment()
	}

	sa.EndWithResult("done")

	return nil
}
