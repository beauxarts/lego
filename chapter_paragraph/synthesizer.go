package chapter_paragraph

import (
	"bytes"
	"fmt"
	gti "github.com/beauxarts/google_tts_integration"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	filenamePaddedDigits  = "09"
	chapterTitleBreakTime = "1s"
	chapterTitlesFilename = "_chapter_titles.txt"
)

func RelChapterTitlesFilename() string {
	return chapterTitlesFilename
}

func RelChapterFilename(chapter int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d"+gti.DefaultEncodingExt, chapter)
}

func RelChapterFfmpegOutputFilename(chapter int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d_ffmpeg.txt", chapter)
}

func RelChapterTitleFilename(chapter int) string {
	return RelChapterParagraphFilename(chapter, 0)
}

func RelChapterParagraphFilename(chapter, paragraph int) string {
	return fmt.Sprintf(
		"%"+filenamePaddedDigits+"d-%"+filenamePaddedDigits+"d"+gti.DefaultEncodingExt,
		chapter,
		paragraph)
}

func RelChapterFilesFilename(chapter int) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d_files.txt", chapter)
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
	overwrite bool) (*Synthesizer, error) {

	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDirectory, 0755); err != nil {
			return nil, err
		}
	}

	return &Synthesizer{
		outputDirectory: outputDirectory,
		synthesizer:     gti.NewSynthesizer(hc, voice, key),
		overwrite:       overwrite,
	}, nil
}

func (s *Synthesizer) CreateChapterTitle(chapter int, content string) error {

	absChapterFilename := filepath.Join(
		s.outputDirectory,
		RelChapterTitleFilename(chapter+1))

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
		RelChapterParagraphFilename(chapter+1, paragraph+1))

	if !s.overwrite {
		if _, err := os.Stat(absChapterParagraphFilename); err == nil {
			return nil
		}
	}

	return s.createContent(content, gti.Text, absChapterParagraphFilename)
}

func (s *Synthesizer) createContent(
	content string,
	contentType gti.SynthesisInputType,
	outputFilename string) error {

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
	defer oggFile.Close()
	if err != nil {
		return err
	}

	if _, err = io.Copy(oggFile, bytes.NewReader(bts)); err != nil {
		return err
	}

	return nil
}

func (s *Synthesizer) CreateChapterFilesList(chapter, paragraphsCount int) error {

	cfn := filepath.Join(
		s.outputDirectory,
		RelChapterFilesFilename(chapter+1))

	chaptersFile, err := os.Create(cfn)
	defer chaptersFile.Close()
	if err != nil {
		return err
	}

	for pp := -1; pp < paragraphsCount; pp++ {
		fn := RelChapterParagraphFilename(chapter+1, pp+1)
		if _, err = io.WriteString(chaptersFile, fmt.Sprintf("file '%s'\n", fn)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Synthesizer) CreateChapterTitles(chapterTitles []string) error {
	ctfn := filepath.Join(
		s.outputDirectory,
		chapterTitlesFilename)

	chapterTitlesFile, err := os.Create(ctfn)
	defer chapterTitlesFile.Close()
	if err != nil {
		return err
	}

	if _, err := io.WriteString(chapterTitlesFile, strings.Join(chapterTitles, "\n")); err != nil {
		return err
	}

	return nil
}
