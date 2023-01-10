package cli

import (
	"bufio"
	"fmt"
	"github.com/beauxarts/lego/chapter_paragraph"
	"github.com/beauxarts/lego/ffmpeg_integration"
	"github.com/boggydigital/nod"
	"github.com/boggydigital/wits"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func ChapterMetadataHandler(u *url.URL) error {
	q := u.Query()

	directory := q.Get("directory")
	importMetadata := q.Get("import-metadata")
	title, author := q.Get("title"), q.Get("author")

	overwrite := q.Has("overwrite")

	return ChapterMetadata(directory, importMetadata, title, author, overwrite)
}

func ChapterMetadata(directory, importMetadata, title, author string, overwrite bool) error {
	cma := nod.Begin("generating ffmpeg chapter metadata...")
	defer cma.End()

	mfn := filepath.Join(
		directory,
		ffmpeg_integration.MetadataFilename)

	if !overwrite {
		if _, err := os.Stat(mfn); err == nil {
			cma.EndWithResult("metadata already exist")
			return nil
		}
	}

	ctfn := filepath.Join(directory, chapter_paragraph.RelChaptersFilename())
	ctf, err := os.Open(ctfn)
	defer ctf.Close()
	if err != nil {
		return cma.EndWithError(err)
	}

	chaptersFileTitle := make(map[string]string, 0)
	chaptersFileDuration := make(map[string]int64)

	scanner := bufio.NewScanner(ctf)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "=")
		if len(parts) < 2 {
			continue
		}
		fn, ct := parts[0], parts[1]
		chaptersFileTitle[fn] = ct
	}

	for cfn := range chaptersFileTitle {

		fofn := filepath.Join(
			directory,
			chapter_paragraph.RelChapterFfmpegOutputFilename(cfn))

		dur, err := ffmpeg_integration.ExtractChapterDuration(fofn)
		if err != nil {
			return cma.EndWithError(err)
		}

		chaptersFileDuration[cfn] = dur
	}

	metadata := make(map[string]string)

	if _, err := os.Stat(importMetadata); err == nil {
		imf, err := os.Open(importMetadata)
		if err != nil {
			return cma.EndWithError(err)
		}
		skv, err := wits.ReadSectionKeyValue(imf)
		if err != nil {
			return cma.EndWithError(err)
		}

		if len(skv) > 0 {
			for _, kv := range skv {
				metadata = kv
				break
			}
		}
	}

	if title != "" {
		metadata["title"] = title
	}
	if author != "" {
		metadata["author"] = author
	}

	fmt.Println(chaptersFileDuration)

	if err := ffmpeg_integration.CreateMetadata(mfn, metadata, chaptersFileTitle, chaptersFileDuration); err != nil {
		return cma.EndWithError(err)
	}

	cma.EndWithResult("done")

	return nil
}
