package cli

import (
	"bufio"
	"errors"
	"fmt"
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
	defer bca.Done()

	absChaptersFilename := filepath.Join(directory, relChaptersFilename())
	chaptersFile, err := os.Open(absChaptersFilename)
	if err != nil {
		return err
	}
	defer chaptersFile.Close()

	relChapterFilenames := make([]string, 0)

	scanner := bufio.NewScanner(chaptersFile)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "=")
		if len(parts) < 1 {
			continue
		}
		relChapterFilenames = append(relChapterFilenames, parts[0])
	}

	bca.TotalInt(len(relChapterFilenames))

	absBookFilesFilename := filepath.Join(directory, relBookFilesFilename())

	bookFilesFile, err := os.Create(absBookFilesFilename)
	if err != nil {
		return err
	}
	defer bookFilesFile.Close()

	for _, relCfn := range relChapterFilenames {

		if err = bindChapter(directory, relCfn, bookFilesFile, ffmpegCmd, overwrite); err != nil {
			return err
		}

		bca.Increment()
	}

	return nil
}

func bindChapter(directory, relChapterFilename string, bookFilesFile io.Writer, ffmpegCmd string, overwrite bool) error {
	absCfn := filepath.Join(directory, relChapterFilename)

	relChapterFoFilename := relChapterFfmpegOutputFilename(absCfn)

	if _, err := os.Stat(absCfn); err == nil {
		if !overwrite {
			return nil
		} else {
			if err := os.Remove(absCfn); err != nil {
				return err
			}
		}
	}

	relChFiFilename := relChapterFilesFilename(absCfn)

	chapterFoFile, err := os.Create(relChapterFoFilename)
	if err != nil {
		return err
	}
	defer chapterFoFile.Close()

	args := []string{"-f", "concat", "-i", relChFiFilename, "-c", "copy", absCfn}
	cmd := exec.Command(ffmpegCmd, args...)
	cmd.Stdout = chapterFoFile
	cmd.Stderr = chapterFoFile
	if err = cmd.Run(); err != nil {
		return err
	}

	fileLine := fmt.Sprintf("file '%s'\n", relChapterFilename)
	if _, err = io.WriteString(bookFilesFile, fileLine); err != nil {
		return err
	}

	return nil
}
