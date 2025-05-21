package cli

import (
	"bufio"
	"github.com/beauxarts/binder/ffmpeg_integration"
	"github.com/boggydigital/nod"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func ChapterMetadataHandler(u *url.URL) error {
	q := u.Query()

	directory := q.Get("directory")
	title, author := q.Get("title"), q.Get("author")

	overwrite := q.Has("overwrite")

	return ChapterMetadata(directory, title, author, overwrite)
}

func ChapterMetadata(directory, title, author string, overwrite bool) error {
	cma := nod.Begin("generating ffmpeg chapter metadata...")
	defer cma.Done()

	mfn := filepath.Join(
		directory,
		ffmpeg_integration.MetadataFilename)

	if !overwrite {
		if _, err := os.Stat(mfn); err == nil {
			cma.EndWithResult("metadata already exist")
			return nil
		}
	}

	absChaptersFilename := filepath.Join(directory, relChaptersFilename())
	chaptersFile, err := os.Open(absChaptersFilename)
	if err != nil {
		return err
	}
	defer chaptersFile.Close()

	chaptersFileTitle := make(map[string]string)
	chaptersFileDuration := make(map[string]int64)

	scanner := bufio.NewScanner(chaptersFile)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "=")
		if len(parts) < 2 {
			continue
		}
		fn, ct := parts[0], parts[1]
		chaptersFileTitle[fn] = ct
	}

	for cfn := range chaptersFileTitle {

		relChapterFoFilename := filepath.Join(directory, relChapterFfmpegOutputFilename(cfn))

		dur, err := ffmpeg_integration.ExtractChapterDuration(relChapterFoFilename)
		if err != nil {
			return err
		}

		chaptersFileDuration[cfn] = dur
	}

	if err := ffmpeg_integration.CreateMetadata(mfn, title, author, chaptersFileTitle, chaptersFileDuration); err != nil {
		return err
	}

	return nil
}
