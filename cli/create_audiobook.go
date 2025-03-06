package cli

import (
	"errors"
	"github.com/boggydigital/nod"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func CreateAudiobookHandler(u *url.URL) error {

	q := u.Query()

	textFilename := q.Get("text-filename")
	outputDirectory := q.Get("output-directory")

	provider := q.Get("provider")
	region := q.Get("region")

	key := q.Get("key-value")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		}
	}

	voiceParams := strings.Split(q.Get("voice-params"), ",")

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

	return CreateAudiobook(
		textFilename,
		outputDirectory,
		provider, region, key,
		voiceParams,
		importMetadata, title, author,
		coverFilename,
		ffmpegCmd, mp4artCmd,
		overwrite)
}

func CreateAudiobook(
	textFilename string,
	outputDirectory string,
	provider, reqion, key string,
	voiceParams []string,
	importMetadata, title, author string,
	coverFilename string,
	ffmpegCmd, mp4artCmd string,
	overwrite bool) error {

	//create audiobook =
	//synthesize +
	//bind-chapters +
	//chapter-metadata +
	//bind-book +
	//cover

	caa := nod.Begin("creating audiobook...")
	defer caa.Done()

	if err := Synthesize(textFilename, outputDirectory, provider, reqion, key, voiceParams, overwrite); err != nil {
		return err
	}

	if err := BindChapters(outputDirectory, ffmpegCmd, overwrite); err != nil {
		return err
	}

	if err := ChapterMetadata(outputDirectory, importMetadata, title, author, overwrite); err != nil {
		return err
	}

	bookFilename, err := BindBook(outputDirectory, ffmpegCmd, overwrite)
	if err != nil {
		return err
	}

	if mp4artCmd != "" && coverFilename != "" {
		if err = Cover(bookFilename, coverFilename, mp4artCmd); err != nil {
			return err
		}
	}

	return nil
}
