# binder

## Installing and running binder

Assuming you have [Go installed](https://go.dev/doc/install), run `go install github.com/beauxarts/binder@latest` and you'll get `binder` binary in your `$GOPATH/bin`.

## Packing a M4B audiobook from individual MP3 files with chapters and a cover

Run the following command:

```bash
binder pack-audiobook -directory PATH-TO-DIRECTORY-WITH-MP3S -title BOOK-TITLE -author BOOK-AUTHOR -cover PATH-TO-COVER-FILENAME
```

Packing an audiobook requires `ffmpeg` to bind individual audio files into a single .mp4 file. Optionally, if cover image is provided, `mp4art` (included in the `mp4v2` tools) is needed. In most cases you don't need to specify the path to those binaries, however you can with the `ffmpeg-cmd` and `mp4art-cmd` parameters.
