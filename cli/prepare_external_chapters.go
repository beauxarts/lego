package cli

import (
	"github.com/beauxarts/lego/chapter_paragraph"
	"github.com/boggydigital/nod"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func PrepareExternalChaptersHandler(u *url.URL) error {
	q := u.Query()

	directory := q.Get("directory")
	extension := q.Get("extension")

	return PrepareExternalChapters(directory, extension)
}

func PrepareExternalChapters(directory, ext string) error {

	pca := nod.NewProgress("preparing chapters...")
	defer pca.Done()

	di, err := os.Open(directory)
	if err != nil {
		return err
	}

	existingFiles, err := di.Readdirnames(-1)
	if err != nil {
		return err
	}

	sort.Strings(existingFiles)
	chapter := 0

	pca.TotalInt(len(existingFiles))

	for _, fn := range existingFiles {

		if !strings.HasSuffix(fn, ext) {
			continue
		}

		relNew := chapter_paragraph.RelChapterParagraphFilename(chapter+1, 0, ext)

		absOld := filepath.Join(directory, fn)
		absNew := filepath.Join(directory, relNew)

		if err := os.Rename(absOld, absNew); err != nil {
			return err
		}

		if err := chapter_paragraph.CreateChapterFilesList(directory, chapter, 0, ext); err != nil {
			return err
		}

		chapter++
		pca.Increment()
	}

	chapterTitles := make([]string, chapter)
	for c := 1; c <= chapter; c++ {
		chapterTitles[c-1] = strconv.Itoa(c)
	}

	if err := chapter_paragraph.CreateChapters(directory, ext, chapterTitles); err != nil {
		return err
	}

	return nil
}
