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
	"strings"
)

func BindChaptersHandler(u *url.URL) error {
	q := u.Query()

	directory := q.Get("directory")

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

	return BindChapters(directory, ffmpegCmd, overwrite)
}

func BindChapters(directory, ffmpegCmd string, overwrite bool) error {

	bca := nod.NewProgress("binding paragraphs into chapters...")
	defer bca.End()

	mfn := filepath.Join(directory, chapter_paragraph.RelChaptersFilename())
	mf, err := os.Open(mfn)
	defer mf.Close()
	if err != nil {
		return bca.EndWithError(err)
	}

	chapterFiles := make([]string, 0)

	scanner := bufio.NewScanner(mf)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "=")
		if len(parts) < 1 {
			continue
		}
		chapterFiles = append(chapterFiles, parts[0])
	}

	bca.TotalInt(len(chapterFiles))

	bfn := filepath.Join(
		directory,
		chapter_paragraph.RelBookFilesFilename())

	bf, err := os.Create(bfn)
	defer bf.Close()
	if err != nil {
		return bca.EndWithError(err)
	}

	for _, relCfn := range chapterFiles {

		absCfn := filepath.Join(directory, relCfn)

		cbofn := chapter_paragraph.RelChapterFfmpegOutputFilename(absCfn)

		if _, err = os.Stat(absCfn); err == nil {
			if !overwrite {
				continue
			} else {
				if err := os.Remove(absCfn); err != nil {
					return bca.EndWithError(err)
				}
			}

		}

		cflfn := chapter_paragraph.RelChapterFilesFilename(absCfn)

		cbof, err := os.Create(cbofn)
		defer cbof.Close()
		if err != nil {
			return bca.EndWithError(err)
		}

		args := []string{"-f", "concat", "-i", cflfn, "-c", "copy", absCfn}
		cmd := exec.Command(ffmpegCmd, args...)
		cmd.Stdout = cbof
		cmd.Stderr = cbof
		if err = cmd.Run(); err != nil {
			return bca.EndWithError(err)
		}

		fileLine := fmt.Sprintf("file '%s'\n", relCfn)
		if _, err = io.WriteString(bf, fileLine); err != nil {
			return bca.EndWithError(err)
		}

		bca.Increment()
	}

	bca.EndWithResult("done")

	return nil
}
