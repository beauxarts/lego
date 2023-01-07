package cli

import (
	"bufio"
	"github.com/beauxarts/lego/chapter_paragraph"
	"github.com/beauxarts/lego/ffmpeg_integration"
	"github.com/boggydigital/nod"
	"net/url"
	"os"
	"path/filepath"
)

func ChapterMetadataHandler(u *url.URL) error {
	q := u.Query()

	inputDirectory := q.Get("input-directory")
	title, author := q.Get("title"), q.Get("author")

	overwrite := q.Has("overwrite")

	return ChapterMetadata(inputDirectory, title, author, overwrite)
}

func ChapterMetadata(inputDirectory, title, author string, overwrite bool) error {
	cma := nod.Begin("generating ffmpeg chapter metadata...")
	defer cma.End()

	mfn := filepath.Join(
		inputDirectory,
		ffmpeg_integration.MetadataFilename)

	if !overwrite {
		if _, err := os.Stat(mfn); err == nil {
			cma.EndWithResult("metadata already exist")
			return nil
		}
	}

	ctfn := filepath.Join(inputDirectory, chapter_paragraph.RelChapterTitlesFilename())
	ctf, err := os.Open(ctfn)
	defer ctf.Close()
	if err != nil {
		return cma.EndWithError(err)
	}

	chapters := make([]string, 0)
	chaptersDuration := make(map[string]int64)

	scanner := bufio.NewScanner(ctf)
	for scanner.Scan() {
		chapters = append(chapters, scanner.Text())
	}

	for ci, ct := range chapters {

		fofn := filepath.Join(
			inputDirectory,
			chapter_paragraph.RelChapterFfmpegOutputFilename(ci+1))

		dur, err := ffmpeg_integration.ExtractChapterDuration(fofn)
		if err != nil {
			return cma.EndWithError(err)
		}

		chaptersDuration[ct] = dur
	}

	if err := ffmpeg_integration.CreateMetadata(mfn, title, author, chapters, chaptersDuration); err != nil {
		return cma.EndWithError(err)
	}

	cma.EndWithResult("done")

	return nil
}
