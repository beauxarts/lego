package chapter_paragraph

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	filenamePaddedDigits = "09"
	listExt              = ".txt"
	filesSuffix          = "_files"
	ffmpegSuffix         = "_ffmpeg"
	bookPrefix           = "_book"
)

func RelChaptersFilename() string {
	return chaptersFilename
}

func RelChapterFilename(chapter int, ext string) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d"+ext, chapter)
}

func RelChapterFfmpegOutputFilename(chapterFilename string) string {
	chapterFilename = strings.TrimSuffix(chapterFilename, filepath.Ext(chapterFilename))
	return fmt.Sprintf(chapterFilename + ffmpegSuffix + listExt)
}

func RelChapterTitleFilename(chapter int, ext string) string {
	return RelChapterParagraphFilename(chapter, 0, ext)
}

func RelChapterParagraphFilename(chapter, paragraph int, ext string) string {
	return fmt.Sprintf(
		"%"+filenamePaddedDigits+"d-%"+filenamePaddedDigits+"d"+ext,
		chapter,
		paragraph)
}

func RelChapterFilesFilename(chapterFilename string) string {
	chapterFilename = strings.TrimSuffix(chapterFilename, filepath.Ext(chapterFilename))
	return fmt.Sprintf(chapterFilename + filesSuffix + listExt)
}

func RelBookFilesFilename() string {
	return bookPrefix + filesSuffix + listExt
}
