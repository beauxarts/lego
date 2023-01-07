package cli

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/beauxarts/lego/chapter_paragraph"
	"github.com/boggydigital/nod"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func BindChaptersHandler(u *url.URL) error {
	q := u.Query()

	inputDirectory := q.Get("input-directory")

	ffmpegCmd := q.Get("ffmpeg-cmd")
	if ffmpegCmd == "" {
		if path, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegCmd = path
		}
	}

	if ffmpegCmd == "" {
		return errors.New("binding chapters requires ffmpeg")
	}

	overwrite := q.Has("overwrite")

	return BindChapters(inputDirectory, ffmpegCmd, overwrite)
}

func BindChapters(inputDirectory, ffmpegCmd string, overwrite bool) error {

	bca := nod.NewProgress("binding paragraphs into chapters...")
	defer bca.End()

	mfn := filepath.Join(inputDirectory, chapter_paragraph.RelChapterTitlesFilename())
	mf, err := os.Open(mfn)
	defer mf.Close()
	if err != nil {
		return bca.EndWithError(err)
	}

	chapters := 0

	scanner := bufio.NewScanner(mf)
	for scanner.Scan() {
		chapters++
	}

	bca.TotalInt(chapters)

	bfn := filepath.Join(
		inputDirectory,
		chapter_paragraph.RelBookFilesFilename())
	bf, err := os.Create(bfn)
	defer bf.Close()
	if err != nil {
		return bca.EndWithError(err)
	}

	for c := 1; c <= chapters; c++ {
		cfn := filepath.Join(
			inputDirectory,
			chapter_paragraph.RelChapterFilename(c))

		cbofn := filepath.Join(
			inputDirectory,
			chapter_paragraph.RelChapterFfmpegOutputFilename(c))

		if _, err = os.Stat(cfn); err == nil {
			if !overwrite {
				continue
			} else {
				if err := os.Remove(cfn); err != nil {
					return bca.EndWithError(err)
				}
			}

		}

		cflfn := filepath.Join(
			inputDirectory,
			chapter_paragraph.RelChapterFilesFilename(c))

		cbof, err := os.Create(cbofn)
		defer cbof.Close()
		if err != nil {
			return bca.EndWithError(err)
		}

		args := []string{"-f", "concat", "-i", cflfn, "-c", "copy", cfn}
		cmd := exec.Command(ffmpegCmd, args...)
		cmd.Stdout = cbof
		cmd.Stderr = cbof
		if err = cmd.Run(); err != nil {
			return bca.EndWithError(err)
		}

		fileLine := fmt.Sprintf("file '%s'\n", chapter_paragraph.RelChapterFilename(c))
		if _, err = io.WriteString(bf, fileLine); err != nil {
			return bca.EndWithError(err)
		}

		bca.Increment()
	}

	bca.EndWithResult("done")

	return nil
}
