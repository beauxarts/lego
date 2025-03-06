package cli

import (
	"fmt"
	"github.com/boggydigital/nod"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func TranslateHandler(u *url.URL) error {
	q := u.Query()

	filename := q.Get("filename")
	from, to := q.Get("from"), q.Get("to")

	key := q.Get("key-value")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		}
	}

	provider := q.Get("provider")

	return Translate(filename, provider, from, to, key)
}

func Translate(filename, provider, from, to, key string) error {

	_, relFilename := filepath.Split(filename)

	ta := nod.Begin("translating %s...", relFilename)
	defer ta.Done()

	tempDir := filepath.Join(os.TempDir(), strings.TrimSuffix(relFilename, filepath.Ext(relFilename)))

	if err := unzipEpub(filename, tempDir); err != nil {
		return err
	}

	if err := translateDirectory(tempDir, provider, from, to, key); err != nil {
		return err
	}

	ext := filepath.Ext(relFilename)
	tFilename := fmt.Sprintf("%s_%s%s", strings.TrimSuffix(relFilename, ext), to, ext)

	if err := zipEpub(tempDir, tFilename); err != nil {
		return err
	}

	return nil
}
