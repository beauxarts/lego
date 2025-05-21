package cli

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/beauxarts/binder/ffmpeg_integration"
	"github.com/boggydigital/nod"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BindBookHandler(u *url.URL) error {
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

	_, err := BindBook(directory, ffmpegCmd, overwrite)

	return err
}

func BindBook(directory, ffmpegCmd string, overwrite bool) (string, error) {

	bba := nod.Begin("binding chapters into a book...")
	defer bba.Done()

	mfn := filepath.Join(directory, ffmpeg_integration.MetadataFilename)
	if _, err := os.Stat(mfn); os.IsNotExist(err) {
		return "", errors.New("required metadata is not found in the provided directory")
	}

	mf, err := os.Open(mfn)
	defer mf.Close()
	if err != nil {
		return "", err
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

	bfn := filepath.Join(directory, fmt.Sprintf("%s - %s.m4b", author, title))

	if _, err := os.Stat(bfn); err == nil {
		if !overwrite {
			bba.EndWithResult("book already exists")
			return bfn, nil
		} else {
			if err := os.Remove(bfn); err != nil {
				return bfn, err
			}
		}
	}

	absBookFilesFilename := filepath.Join(directory, relBookFilesFilename())

	args := []string{"-f", "concat", "-i", absBookFilesFilename, "-i", mfn, "-map_metadata", "1", bfn}

	cmd := exec.Command(ffmpegCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return bfn, err
	}

	return bfn, nil
}
