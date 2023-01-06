package chapter_paragraph

import (
	"bytes"
	"fmt"
	gti "github.com/beauxarts/google_tts_integration"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	filenamePaddedDigits  = "09"
	chapterTitleBreakTime = "1s"
)

func RelChapterFilename(chapter int) string {
	return RelChapterParagraphFilename(chapter, -1)
}

func RelChapterParagraphFilename(chapter, paragraph int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d-%"+filenamePaddedDigits+"d"+gti.DefaultEncodingExt, chapter+1, paragraph+1)
}

func RelChapterFilesListFilename(chapter int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d.txt", chapter+1)
}

type Synthesizer struct {
	outputDirectory string
	synthesizer     *gti.Synthesizer
	overwrite       bool
}

func NewSynthesizer(
	hc *http.Client,
	voice *gti.VoiceSelectionParams,
	key string,
	outputDirectory string,
	overwrite bool) *Synthesizer {
	return &Synthesizer{
		outputDirectory: outputDirectory,
		synthesizer:     gti.NewSynthesizer(hc, voice, key),
		overwrite:       overwrite,
	}
}

func (s *Synthesizer) CreateChapterTitle(chapter int, content string) error {

	absChapterFilename := filepath.Join(
		s.outputDirectory,
		RelChapterFilename(chapter))

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

	return s.createContent(content, gti.SSML, absChapterFilename)
}

func (s *Synthesizer) CreateChapterParagraph(chapter, paragraph int, content string) error {

	absChapterParagraphFilename := filepath.Join(
		s.outputDirectory,
		RelChapterParagraphFilename(chapter, paragraph))

	if !s.overwrite {
		if _, err := os.Stat(absChapterParagraphFilename); err == nil {
			return nil
		}
	}

	return s.createContent(content, gti.Text, absChapterParagraphFilename)
}

func (s *Synthesizer) createContent(content string, contentType gti.SynthesisInputType, outputFilename string) error {

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

func (s *Synthesizer) CreateChapterFilesList(chapter, paragraphsCount int) error {

	cfn := filepath.Join(
		s.outputDirectory,
		RelChapterFilesListFilename(chapter))

	chaptersFile, err := os.Create(cfn)
	if err != nil {
		return err
	}
	defer chaptersFile.Close()

	for pp := -1; pp < paragraphsCount; pp++ {
		fn := RelChapterParagraphFilename(chapter, pp)
		if _, err = io.WriteString(chaptersFile, fmt.Sprintf("file '%s'\n", fn)); err != nil {
			return err
		}
	}

	return nil
}
