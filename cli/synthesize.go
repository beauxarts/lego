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
	"os/exec"
)

func SynthesizeHandler(u *url.URL) error {

	q := u.Query()

	filename := q.Get("filename")
	key := q.Get("key")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		} else {
			return err
		}
	}

	if key == "" {
		return errors.New("synthesis requires a key as a value or a file")
	}

	vl, vn, vg := q.Get("voice-locale"), q.Get("voice-name"), q.Get("voice-gender")
	voice := gti.NewVoice(vl, vn, vg)

	ffmpegCmd := q.Get("ffmpeg-cmd")
	if ffmpegCmd == "" {
		if path, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegCmd = path
		}
	}

	return Synthesize(filename, voice, key, ffmpegCmd)
}

func Synthesize(filename string, voice *gti.VoiceSelectionParams, key, ffmpegCmd string) error {
	sa := nod.NewProgress("synthesizing audiobook from text...")
	defer sa.End()

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	td := divido.NewTextDocument(file)
	chapters := td.ChapterTitles()

	sa.TotalInt(len(chapters))

	for ci, ct := range chapters {

		if err := synthesizeChapterParagraph(ci, -1, ct, voice, key); err != nil {
			_ = sa.EndWithError(err)
		}

		pa := nod.NewProgress(" synthesizing chapter %d paragraphs...", ci+1)

		paragraphs := td.ChapterParagraphs(ct)
		pa.TotalInt(len(paragraphs))

		for pi, pt := range paragraphs {
			if err := synthesizeChapterParagraph(ci, pi, string(pt), voice, key); err != nil {
				_ = pa.EndWithError(err)
			}
			pa.Increment()
		}

		pa.EndWithResult("done")

		if ffmpegCmd != "" {

			if err := writeChapterFilesList(ci, len(paragraphs)); err != nil {
				panic(err)
			}

			ma := nod.Begin(" merging ogg files into mp3...")

			mp3fn := fmt.Sprintf("%09d.mp3", ci+1)

			if _, err := os.Stat(mp3fn); os.IsNotExist(err) {
				args := []string{"-f", "concat", "-i", fmt.Sprintf("%09d.txt", ci+1), mp3fn}
				cmd := exec.Command(ffmpegCmd, args...)
				if err := cmd.Run(); err != nil {
					_ = ma.EndWithError(err)
				}
			}

			ma.EndWithResult("done")
		}

		//break
		sa.Increment()
	}

	sa.EndWithResult("done")

	return nil
}

func synthesizeChapterParagraph(chapter, paragraph int, text string, voice *gti.VoiceSelectionParams, key string) error {

	ofn := fmt.Sprintf("%09d-%09d.ogg", chapter+1, paragraph+1)
	if _, err := os.Stat(ofn); err == nil {
		return nil
	}

	sr, err := gti.PostTextSynthesize(http.DefaultClient, text, voice, key)
	if err != nil {
		return err
	}

	bts, err := sr.Bytes()
	if err != nil {
		return err
	}

	oggFile, err := os.Create(ofn)
	if err != nil {
		return err
	}

	if _, err = io.Copy(oggFile, bytes.NewReader(bts)); err != nil {
		return err
	}

	return nil
}

func writeChapterFilesList(chapter, paragraphsCount int) error {
	cfn := fmt.Sprintf("%09d.txt", chapter+1)
	chaptersFile, err := os.Create(cfn)
	if err != nil {
		return err
	}
	defer chaptersFile.Close()

	for pp := 0; pp <= paragraphsCount; pp++ {
		if _, err := io.WriteString(chaptersFile, fmt.Sprintf("file '%09d-%09d.ogg'\n", chapter+1, pp)); err != nil {
			return err
		}
	}

	return nil
}
