package chapter_paragraph

import (
	"errors"
	"fmt"
	"github.com/beauxarts/tts_integration"
	"github.com/beauxarts/tts_integration/gcp"
	"github.com/beauxarts/tts_integration/say"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultPauseDuration = time.Second
	chaptersFilename     = "_chapters.txt"
)

type Synthesizer struct {
	outputDirectory string
	synthesizer     tts_integration.Synthesizer
	overwrite       bool
	ext             string
}

func NewSaySynthesizer(
	voiceParams []string,
	outputDirectory string,
	overwrite bool) (*Synthesizer, error) {

	if outputDirectory != "" {
		if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
			if err := os.MkdirAll(outputDirectory, 0755); err != nil {
				return nil, err
			}
		}
	}

	voice := ""
	if len(voiceParams) > 0 {
		voice = voiceParams[0]
	}

	return &Synthesizer{
		outputDirectory: outputDirectory,
		synthesizer:     say.NewSynthesizer(voice, say.DefaultAudioFormat),
		overwrite:       overwrite,
		ext:             say.DefaultAudioExt,
	}, nil
}

func NewGCPSynthesizer(
	hc *http.Client,
	voiceParams []string,
	key string,
	outputDirectory string,
	overwrite bool) (*Synthesizer, error) {

	if outputDirectory != "" {
		if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
			if err := os.MkdirAll(outputDirectory, 0755); err != nil {
				return nil, err
			}
		}
	}

	return &Synthesizer{
		outputDirectory: outputDirectory,
		synthesizer:     gcp.NewSynthesizer(hc, key, voiceParams...),
		overwrite:       overwrite,
		ext:             gcp.DefaultAudioEncodingExt,
	}, nil
}

func (s *Synthesizer) CreateChapterTitle(chapter int, text string) error {

	absChapterFilename := filepath.Join(
		s.outputDirectory,
		RelChapterTitleFilename(chapter+1, s.ext))

	if !s.overwrite {
		if _, err := os.Stat(absChapterFilename); err == nil {
			return nil
		}
	}

	content, contentType := s.synthesizer.DecorateWithPauses(text, defaultPauseDuration)

	return s.createContent(content, contentType, absChapterFilename)
}

func (s *Synthesizer) CreateChapterParagraph(chapter, paragraph int, content string) error {

	absChapterParagraphFilename := filepath.Join(
		s.outputDirectory,
		RelChapterParagraphFilename(chapter+1, paragraph+1, s.ext))

	if !s.overwrite {
		if _, err := os.Stat(absChapterParagraphFilename); err == nil {
			return nil
		}
	}

	return s.createContent(content, tts_integration.Text, absChapterParagraphFilename)
}

func (s *Synthesizer) CreatePause(chapter, paragraph int) error {

	absChapterParagraphFilename := filepath.Join(
		s.outputDirectory,
		RelChapterParagraphFilename(chapter+1, paragraph+1, s.ext))

	if !s.overwrite {
		if _, err := os.Stat(absChapterParagraphFilename); err == nil {
			return nil
		}
	}

	content, contentType := s.synthesizer.DecorateWithPauses("", defaultPauseDuration)

	return s.createContent(content, contentType, absChapterParagraphFilename)
}

func (s *Synthesizer) createContent(
	content string,
	contentType tts_integration.SynthesisInputType,
	outputFilename string) error {

	var writer *os.File
	var err error
	if s.synthesizer.IsWriterRequired() {
		writer, err = os.Create(outputFilename)
		defer writer.Close()
		if err != nil {
			return err
		}
	}

	if contentType == tts_integration.SSML &&
		!s.synthesizer.IsSSMLSupported() {
		return errors.New("synthesizer doesn't support SSML")
	}

	switch contentType {
	case tts_integration.Text:
		return s.synthesizer.WriteText(content, writer, outputFilename)
	case tts_integration.SSML:
		return s.synthesizer.WriteSSML(content, writer, outputFilename)
	}

	return errors.New("unsupported content type " + contentType.String())
}

func (s *Synthesizer) CreateChapterFilesList(chapter, paragraphsCount int) error {

	chapterFilename := RelChapterFilename(chapter+1, s.ext)

	cfn := filepath.Join(
		s.outputDirectory,
		RelChapterFilesFilename(chapterFilename))

	chaptersFile, err := os.Create(cfn)
	defer chaptersFile.Close()
	if err != nil {
		return err
	}

	for pp := -1; pp < paragraphsCount; pp++ {
		fn := RelChapterParagraphFilename(chapter+1, pp+1, s.ext)
		if _, err = io.WriteString(chaptersFile, fmt.Sprintf("file '%s'\n", fn)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Synthesizer) CreateChapters(chapterTitles []string) error {
	ctfn := filepath.Join(
		s.outputDirectory,
		chaptersFilename)

	chapterTitlesFile, err := os.Create(ctfn)
	defer chapterTitlesFile.Close()
	if err != nil {
		return err
	}

	for ci, ct := range chapterTitles {
		if _, err := io.WriteString(chapterTitlesFile, fmt.Sprintf("%s=%s\n", RelChapterFilename(ci+1, s.ext), ct)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Synthesizer) Voices(locale string) ([]string, error) {
	return s.synthesizer.VoicesStrings(locale)
}
