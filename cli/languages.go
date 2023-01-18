package cli

import (
	"errors"
	"github.com/beauxarts/polyglot"
	"github.com/beauxarts/polyglot/acs"
	"github.com/beauxarts/polyglot/gcp"
	"github.com/boggydigital/nod"
	"net/http"
	"net/url"
	"os"
)

func LanguagesHandler(u *url.URL) error {
	q := u.Query()

	language := q.Get("language")
	key := q.Get("key-value")
	if key == "" {
		//attempt to get the key from a file, if specified
		keyFilename := q.Get("key-filename")
		if keyBytes, err := os.ReadFile(keyFilename); err == nil {
			key = string(keyBytes)
		}
	}

	provider := q.Get("provider")

	return Languages(provider, language, key)
}

func Languages(provider, language, key string) error {
	la := nod.Begin("languages available for translations:")
	defer la.End()

	var translator polyglot.Translator
	var err error

	switch provider {
	case "gcp":
		translator, err = gcp.NewTranslator(http.DefaultClient, gcp.NeuralMachineTranslation, key)
	case "acs":
		translator, err = acs.NewTranslator(http.DefaultClient, key)
	default:
		err = errors.New("unknown provider " + provider)
	}

	if err != nil {
		return la.EndWithError(err)
	}

	languages, err := translator.Languages(language)
	if err != nil {
		return la.EndWithError(err)
	}

	for lang, name := range languages {
		v := nod.Begin("- %s: %s", lang, name)
		v.End()
	}

	la.EndWithResult("done")

	return nil
}
