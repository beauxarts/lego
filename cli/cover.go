package cli

import (
	"errors"
	"github.com/boggydigital/nod"
	"net/url"
	"os"
	"os/exec"
)

func CoverHandler(u *url.URL) error {
	q := u.Query()

	inputFilename := q.Get("input-filename")
	coverFilename := q.Get("cover-filename")

	mp4artCmd := q.Get("mp4art-cmd")
	if mp4artCmd == "" {
		if path, err := exec.LookPath("mp4art"); err == nil {
			mp4artCmd = path
		}
	}

	if mp4artCmd == "" {
		return errors.New("adding cover requires mp4art (part of mp4v2)")
	}

	return Cover(inputFilename, coverFilename, mp4artCmd)
}

func Cover(inputFilename, coverFilename, mp4artCmd string) error {

	aca := nod.Begin("adding cover image...")
	defer aca.End()

	if _, err := os.Stat(inputFilename); os.IsNotExist(err) {
		return aca.EndWithError(errors.New("input file not found"))
	}

	if _, err := os.Stat(coverFilename); os.IsNotExist(err) {
		return aca.EndWithError(errors.New("cover file not found"))
	}

	args := []string{"--add", coverFilename, inputFilename}
	cmd := exec.Command(mp4artCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return aca.EndWithError(err)
	}

	aca.EndWithResult("done")

	return nil
}
