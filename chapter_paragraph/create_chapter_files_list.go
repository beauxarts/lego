package chapter_paragraph

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CreateChapterFilesList(directory string, chapter, paragraphsCount int, ext string) error {

	chapterFilename := RelChapterFilename(chapter+1, ext)

	cfn := filepath.Join(
		directory,
		RelChapterFilesFilename(chapterFilename))

	chaptersFile, err := os.Create(cfn)
	defer chaptersFile.Close()
	if err != nil {
		return err
	}

	for pp := -1; pp < paragraphsCount; pp++ {
		fn := RelChapterParagraphFilename(chapter+1, pp+1, ext)
		if _, err = io.WriteString(chaptersFile, fmt.Sprintf("file '%s'\n", fn)); err != nil {
			return err
		}
	}

	return nil
}
