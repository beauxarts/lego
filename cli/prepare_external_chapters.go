package cli

import (
	"fmt"
	"github.com/boggydigital/nod"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const chaptersFilename = "_chapters.txt"

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

		relOldChPaFilename := relChapterParagraphFilename(chapter+1, 0, ext)

		absOldChPaFilename := filepath.Join(directory, fn)
		absNewChPaFilename := filepath.Join(directory, relOldChPaFilename)

		if err = os.Rename(absOldChPaFilename, absNewChPaFilename); err != nil {
			return err
		}

		if err = createChapterFilesList(directory, chapter, 0, ext); err != nil {
			return err
		}

		chapter++
		pca.Increment()
	}

	chapterTitles := make([]string, chapter)
	for c := 1; c <= chapter; c++ {
		chapterTitles[c-1] = strconv.Itoa(c)
	}

	if err = createChapters(directory, ext, chapterTitles); err != nil {
		return err
	}

	return nil
}

func createChapters(directory, ext string, chapterTitles []string) error {
	ctfn := filepath.Join(
		directory,
		chaptersFilename)

	chapterTitlesFile, err := os.Create(ctfn)
	if err != nil {
		return err
	}
	defer chapterTitlesFile.Close()

	for ci, ct := range chapterTitles {
		if _, err := io.WriteString(chapterTitlesFile, fmt.Sprintf("%s=%s\n", relChapterFilename(ci+1, ext), ct)); err != nil {
			return err
		}
	}

	return nil
}

func createChapterFilesList(directory string, chapter, paragraphsCount int, ext string) error {

	chapterFilename := relChapterFilename(chapter+1, ext)

	absChFiFilename := filepath.Join(directory, relChapterFilesFilename(chapterFilename))

	chaptersFile, err := os.Create(absChFiFilename)
	if err != nil {
		return err
	}
	defer chaptersFile.Close()

	for pp := -1; pp < paragraphsCount; pp++ {
		relChPaFilename := relChapterParagraphFilename(chapter+1, pp+1, ext)
		if _, err = io.WriteString(chaptersFile, fmt.Sprintf("file '%s'\n", relChPaFilename)); err != nil {
			return err
		}
	}

	return nil
}
