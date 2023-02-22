package chapter_paragraph

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CreateChapters(directory, ext string, chapterTitles []string) error {
	ctfn := filepath.Join(
		directory,
		chaptersFilename)

	chapterTitlesFile, err := os.Create(ctfn)
	defer chapterTitlesFile.Close()
	if err != nil {
		return err
	}

	for ci, ct := range chapterTitles {
		if _, err := io.WriteString(chapterTitlesFile, fmt.Sprintf("%s=%s\n", RelChapterFilename(ci+1, ext), ct)); err != nil {
			return err
		}
	}

	return nil
}
