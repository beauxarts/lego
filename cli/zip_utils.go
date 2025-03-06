package cli

import (
	"archive/zip"
	"github.com/boggydigital/nod"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func unzipEpub(filename, dir string) error {

	_, relFilename := filepath.Split(filename)

	uea := nod.NewProgress(" unpacking %s...", relFilename)
	defer uea.Done()

	zr, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer zr.Close()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	zipFiles := zr.File

	uea.TotalInt(len(zipFiles))

	for _, zipFile := range zipFiles {
		//check for Zip Slip
		if strings.Contains(zipFile.Name, "..") {
			continue
		}
		if err := unzipTo(zipFile, dir); err != nil {
			return err
		}
		uea.Increment()
	}

	return nil
}

func unzipTo(zipFile *zip.File, dir string) error {
	file, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	absFilename := filepath.Join(dir, zipFile.Name)

	if zipFile.FileInfo().IsDir() {
		if err = os.MkdirAll(absFilename, 0755); err != nil {
			return err
		}
	} else {
		if err = os.MkdirAll(filepath.Dir(absFilename), 0755); err != nil {
			return err
		}

		f, err := os.Create(absFilename)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err = io.Copy(f, file); err != nil {
			return err
		}
	}
	return nil
}

func zipEpub(dir, filename string) error {

	_, relFilename := filepath.Split(filename)
	zea := nod.NewProgress(" packing %s...", relFilename)
	defer zea.Done()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	zw := &zipWalker{
		w:    w,
		root: dir,
	}

	err = filepath.Walk(dir, zw.zipPath)
	if err != nil {
		return err
	}

	return nil
}

type zipWalker struct {
	root string
	w    *zip.Writer
}

func (zw *zipWalker) zipPath(path string, i os.FileInfo, err error) error {

	if err != nil {
		return err
	}
	if i.IsDir() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	relPath, err := filepath.Rel(zw.root, path)
	if err != nil {
		return err
	}

	f, err := zw.w.Create(relPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}

	return nil
}
