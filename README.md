# lego

lego - "I read" in Latin.

Also, lego is a lightweight CLI utility that can:

- synthesize audiobooks from .txt files and bind them into .m4b books with chapters and (optional) cover;
- translate EPUB books from one language to another.

## Installing and running lego

Assuming you have [Go installed](https://go.dev/doc/install), run `go install github.com/beauxarts/lego` and you'll get `lego` binary in your `$GOPATH/bin`.

## Synthesizing audiobooks with lego

Lego allows a choice of local and cloud neural voices supported by providers in https://github.com/beauxarts/tts_integration:

- `acs` - Azure Cognitive Services Text to Speech
- `gcp` - Google Cloud Text to Speech
- `say` - (macOS only) any locally installed voices supported by the 'say' command

### Getting the list of supported voices

Run the following command:

```bash
lego voices -locale LOCALE -provider PROVIDER [additional provider specific parameters]
```

Additional parameters for `gcp` provider:

- `key-file` or `key-value` - path to a file containing secret key or literal key value

Additional parameters for `acs` provider:

- `key-file` or `key-value` - path to a file containing secret key or literal key value
- `region` - ACS region for your resource (e.g. `eastus`)

### Creating an audiobook

Run the following command:

```bash
lego create-audiobook -text-filename PATH-TO-TXT-FILE -output-directory PATH-TO-DIRECTORY -provider PROVIDER [provider specific parameters] -voice-params [name, locale, gender] -title TITLE -author AUTHOR -cover-filename PATH-TO-COVER-IMAGE
```

`lego` uses `divido` to break down text files into chapters and paragraphs. Basically the chapter title is separated from paragraphs by at least 3 new lines, and everything with less new lines separators are paragraphs under that chapter. That seems to match text files converted by `calibre` from EPUB. You can validate chapters with `lego info` command.

Creating audiobook requires `ffmpeg` to bind individual audio files into a single .mp4 file. Optionally, if cover image is provided, `mp4art` (included in the `mp4v2` tools) is needed. In most cases you don't need to specify the path to those binaries, however you can with the `ffmpeg-cmd` and `mp4art-cmd` parameters.

Optionally, you can provide all audiobook metadata in a single text file in [wits format](https://github.com/boggydigital/wits). The most important metadata parameters are `author` and `title`, which you can also specify with parameters. Note: `lego` can use metadata exported by [fedorov](https://github.com/beauxarts/fedorov).

Note: `lego` synthesis can be interrupted and continued. When restarting synthesis existing files will be skipped (to avoid paying unnecessary costs). You can force overwriting with an `overwrite` parameter.     

### Other audiobook creation commands

`create-audiobook` is a sequence of individual command that you can use individually. You can use `lego help COMMAND` to learn more about each command. 

Most users don't need that and would be better served by `create-audiobook` command. 

## Translating EPUBs with lego

Lego allows a choice of cloud neural translation models supported by providers in https://github.com/beauxarts/polyglot:

- `acs` - Azure Cognitive Services Text to Speech
- `gcp` - Google Cloud Text to Speech

There doesn't seem to be a local translator available (at least on macOS). When one appears - I'm hoping to support it.

### Getting the list of supported languages

Run the following command:

```bash
lego languages -language LANGUAGE -provider PROVIDER [additional provider specific parameters]
```

Additional parameters for `gcp`, `acs` providers:

- `key-file` or `key-value` - path to a file containing secret key or literal key value
- 
### Translating EPUB

Run the following command:

```bash
lego translate -filename PATH-TO-EPUB-FILE -provider PROVIDER [provider specific parameters] -from LANGUAGE -to LANGUAGE
```

`lego` will unpack EPUB into a temporary directory, translate each XHTML/HTML file in that folder and then pack EPUB back, so you'll get a new file with the target language suffix in the same location, where the original is. For example, when translating `my_great_book.epub` file into `fr` language, you'll get `my_great_book_fr.epub` as a result.

Note: `lego` translation can be interrupted and continued. When restarting translation already translated files (.xhtml.translated or .html.translated) will be skipped (to avoid paying unnecessary costs). If you want to redo those files - delete .translate files and run `translate` command again. 




