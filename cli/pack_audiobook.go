package cli

import (
	"errors"
	"github.com/boggydigital/nod"
	"net/url"
	"os/exec"
	"strings"
)

const defaultExt = ".mp3"

func PackAudiobookHandler(u *url.URL) error {

	q := u.Query()

	directory := q.Get("directory")
	extension := q.Get("extension")

	importMetadata := q.Get("import-metadata")
	title, author := q.Get("title"), q.Get("author")

	ffmpegCmd := q.Get("ffmpeg-cmd")
	if ffmpegCmd == "" {
		if path, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegCmd = path
		}
	}

	if ffmpegCmd == "" {
		return errors.New("binding chapters requires ffmpeg")
	}

	coverFilename := q.Get("cover-filename")

	mp4artCmd := q.Get("mp4art-cmd")
	if mp4artCmd == "" {
		if path, err := exec.LookPath("mp4art"); err == nil {
			mp4artCmd = path
		}
	}

	overwrite := q.Has("overwrite")

	return PackAudioBook(
		directory,
		extension,
		importMetadata, title, author,
		coverFilename,
		ffmpegCmd, mp4artCmd,
		overwrite)
}

func PackAudioBook(
	directory string,
	extension string,
	importMetadata, title, author string,
	coverFilename string,
	ffmpegCmd, mp4artCmd string,
	overwrite bool) error {

	//pac audiobook =
	//prepare external chapters (rename + generate chapter .txt files) +
	//bind-chapters +
	//chapter-metadata +
	//bind-book +
	//cover

	pa := nod.Begin("packing audiobook...")
	defer pa.End()

	if extension == "" {
		extension = defaultExt
	} else if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	if err := PrepareExternalChapters(directory, extension); err != nil {
		return pa.EndWithError(err)
	}

	if err := BindChapters(directory, ffmpegCmd, overwrite); err != nil {
		return pa.EndWithError(err)
	}

	if err := ChapterMetadata(directory, importMetadata, title, author, overwrite); err != nil {
		return pa.EndWithError(err)
	}

	bookFilename, err := BindBook(directory, ffmpegCmd, overwrite)
	if err != nil {
		return pa.EndWithError(err)
	}

	if mp4artCmd != "" && coverFilename != "" {
		if err := Cover(bookFilename, coverFilename, mp4artCmd); err != nil {
			return pa.EndWithError(err)
		}
	}

	pa.EndWithResult("done")

	return nil
}
