package cli

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/beauxarts/divido"
	gti "github.com/beauxarts/google_tts_integration"
	"github.com/boggydigital/nod"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	filenamePaddedDigits  = "09"
	chapterTitleBreakTime = "1s"
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

	ncps := newChapterParagraphSynthesizer(http.DefaultClient, voice, key, outputDirectory, overwrite)

	sa.TotalInt(len(chapters))

	for ci, ct := range chapters {

		if err := ncps.synthesizeChapterTitle(ci, ct); err != nil {
			return sa.EndWithError(err)
		}

		pa := nod.NewProgress(" synthesizing chapter %d paragraphs...", ci+1)

		paragraphs := td.ChapterParagraphs(ct)
		pa.TotalInt(len(paragraphs))

		for pi, pt := range paragraphs {
			if err = ncps.synthesizeChapterParagraph(ci, pi, string(pt)); err != nil {
				return pa.EndWithError(err)
			}
			pa.Increment()
		}

		pa.EndWithResult("done")

		if err = ncps.createChapterFilesList(ci, len(paragraphs)); err != nil {
			return sa.EndWithError(err)
		}

		sa.Increment()
	}

	sa.EndWithResult("done")

	return nil
}

func relChapterFilename(chapter int) string {
	return relChapterParagraphFilename(chapter, -1)
}

func relChapterParagraphFilename(chapter, paragraph int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d-%"+filenamePaddedDigits+"d.ogg", chapter+1, paragraph+1)
}

func relChapterFilesListFilename(chapter int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d.txt", chapter+1)
}

type chapterParagraphSynthesizer struct {
	outputDirectory string
	synthesizer     *gti.Synthesizer
	overwrite       bool
}

func newChapterParagraphSynthesizer(
	hc *http.Client,
	voice *gti.VoiceSelectionParams,
	key string,
	outputDirectory string,
	overwrite bool) *chapterParagraphSynthesizer {
	return &chapterParagraphSynthesizer{
		outputDirectory: outputDirectory,
		synthesizer:     gti.NewSynthesizer(hc, voice, key),
		overwrite:       overwrite,
	}
}

func (s *chapterParagraphSynthesizer) synthesizeChapterTitle(chapter int, content string) error {

	absChapterFilename := filepath.Join(
		s.outputDirectory,
		relChapterFilename(chapter))

	if !s.overwrite {
		if _, err := os.Stat(absChapterFilename); err == nil {
			return nil
		}
	}

	content = fmt.Sprintf(
		"<speak><break time=\"%s\"/>%s<break time=\"%s\"/></speak>",
		chapterTitleBreakTime,
		content,
		chapterTitleBreakTime)

	return s.synthesizeContent(content, gti.SSML, absChapterFilename)
}

func (s *chapterParagraphSynthesizer) synthesizeChapterParagraph(chapter, paragraph int, content string) error {

	absChapterParagraphFilename := filepath.Join(
		s.outputDirectory,
		relChapterParagraphFilename(chapter, paragraph))

	if !s.overwrite {
		if _, err := os.Stat(absChapterParagraphFilename); err == nil {
			return nil
		}
	}

	return s.synthesizeContent(content, gti.Text, absChapterParagraphFilename)
}

func (s *chapterParagraphSynthesizer) synthesizeContent(content string, contentType gti.SynthesisInputType, outputFilename string) error {

	var postContent func(string) (*gti.TextSynthesizeResponse, error)

	switch contentType {
	case gti.Text:
		postContent = s.synthesizer.PostText
	case gti.SSML:
		postContent = s.synthesizer.PostSSML
	}

	sr, err := postContent(content)
	if err != nil {
		return err
	}

	bts, err := sr.Bytes()
	if err != nil {
		return err
	}

	oggFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer oggFile.Close()

	if _, err = io.Copy(oggFile, bytes.NewReader(bts)); err != nil {
		return err
	}

	return nil
}

func (s *chapterParagraphSynthesizer) createChapterFilesList(chapter, paragraphsCount int) error {

	cfn := filepath.Join(
		s.outputDirectory,
		relChapterFilesListFilename(chapter))

	chaptersFile, err := os.Create(cfn)
	if err != nil {
		return err
	}
	defer chaptersFile.Close()

	for pp := -1; pp < paragraphsCount; pp++ {
		fn := relChapterParagraphFilename(chapter, pp)
		if _, err = io.WriteString(chaptersFile, fmt.Sprintf("file '%s'\n", fn)); err != nil {
			return err
		}
	}

	return nil
}
