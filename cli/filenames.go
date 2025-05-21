package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	filenamePaddedDigits = "09"

	listExt = ".txt"

	filesSuffix  = "_files"
	ffmpegSuffix = "_ffmpeg"
	bookPrefix   = "_book"
)

func relChaptersFilename() string {
	return chaptersFilename
}

func relChapterFilename(chapter int, ext string) string {
	return fmt.Sprintf("%"+filenamePaddedDigits+"d"+ext, chapter)
}

func relChapterFfmpegOutputFilename(chapterFilename string) string {
	chapterFilename = strings.TrimSuffix(chapterFilename, filepath.Ext(chapterFilename))
	return fmt.Sprintf(chapterFilename + ffmpegSuffix + listExt)
}

func relChapterTitleFilename(chapter int, ext string) string {
	return relChapterParagraphFilename(chapter, 0, ext)
}

func relChapterParagraphFilename(chapter, paragraph int, ext string) string {
	return fmt.Sprintf(
		"%"+filenamePaddedDigits+"d-%"+filenamePaddedDigits+"d"+ext,
		chapter,
		paragraph)
}

func relChapterFilesFilename(chapterFilename string) string {
	chapterFilename = strings.TrimSuffix(chapterFilename, filepath.Ext(chapterFilename))
	return fmt.Sprintf(chapterFilename + filesSuffix + listExt)
}

func relBookFilesFilename() string {
	return bookPrefix + filesSuffix + listExt
}
