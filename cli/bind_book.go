package cli

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/beauxarts/lego/chapter_paragraph"
	"github.com/beauxarts/lego/ffmpeg_integration"
	"github.com/boggydigital/nod"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BindBookHandler(u *url.URL) error {
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

	return BindBook(inputDirectory, ffmpegCmd, overwrite)
}

func BindBook(inputDirectory, ffmpegCmd string, overwrite bool) error {

	bba := nod.Begin("binding chapters into a book...")
	defer bba.End()

	mfn := filepath.Join(inputDirectory, ffmpeg_integration.MetadataFilename)
	if _, err := os.Stat(mfn); os.IsNotExist(err) {
		return bba.EndWithError(errors.New("required metadata is not found in the provided directory"))
	}

	mf, err := os.Open(mfn)
	defer mf.Close()
	if err != nil {
		return bba.EndWithError(err)
	}

	title, author := "", ""

	scanner := bufio.NewScanner(mf)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ffmpeg_integration.MetadataChapterSection) {
			break
		}
		if strings.HasPrefix(line, ffmpeg_integration.MetadataTitlePrefix) {
			title = strings.TrimPrefix(line, ffmpeg_integration.MetadataTitlePrefix)
		}
		if strings.HasPrefix(line, ffmpeg_integration.MetadataArtistPrefix) {
			author = strings.TrimPrefix(line, ffmpeg_integration.MetadataArtistPrefix)
		}
	}

	if title == "" {
		title = "Untitled"
	}
	if author == "" {
		author = "Anonymous"
	}

	bfn := filepath.Join(inputDirectory, fmt.Sprintf("%s - %s.m4b", author, title))

	if _, err := os.Stat(bfn); err == nil {
		if !overwrite {
			bba.EndWithResult("book already exists")
			return nil
		} else {
			if err := os.Remove(bfn); err != nil {
				return bba.EndWithError(err)
			}
		}
	}

	bflfn := filepath.Join(inputDirectory, chapter_paragraph.RelBookFilesFilename())

	args := []string{"-f", "concat", "-i", bflfn, "-i", mfn, "-map_metadata", "1", bfn}

	cmd := exec.Command(ffmpegCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return bba.EndWithError(err)
	}

	return nil
}
