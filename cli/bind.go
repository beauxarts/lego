package cli

import (
	"net/url"
	"os/exec"
)

func BindHandler(u *url.URL) error {
	q := u.Query()

	ffmpegCmd := q.Get("ffmpeg-cmd")
	if ffmpegCmd == "" {
		if path, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegCmd = path
		}
	}

	return Bind(ffmpegCmd)
}

func Bind(ffmpegCmd string) error {

	//- (currently chapters are saved as mp3s) using 00000000c.txt, combine all paragraph files into a single chapter file 00000000c.ogg
	//- (not implemented) process conversion output to extract chapter length
	//- (not implemented) delete individual paragraph audio files and list of chapter paragraph audio files
	//- (not implemented) generate FFMETADATA file required for audiobook chapter markers using chapter lengths
	//- (not implemented) bind a single file audiobook with chapter metadata
	//- (not implemented) cleanup everything created in the session leaving just the audiobook

	if ffmpegCmd != "" {

		//ma := nod.Begin(" merging ogg files into mp3...")
		//
		//mp3fn := fmt.Sprintf("%09d.mp3", ci+1)
		//
		//if _, err := os.Stat(mp3fn); os.IsNotExist(err) {
		//	args := []string{"-f", "concat", "-i", fmt.Sprintf("%09d.txt", ci+1), mp3fn}
		//	cmd := exec.Command(ffmpegCmd, args...)
		//	if err := cmd.Run(); err != nil {
		//		_ = ma.EndWithError(err)
		//	}
		//}
		//
		//ma.EndWithResult("done")
	}

	return nil
}
