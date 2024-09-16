package hlsClipper

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// FFmpeg - segment merger
func mergeSegments(concatFile, outputFile string) error {
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", tmpPath+concatFile, "-c", "copy", tmpPath+outputFile)
	return cmd.Run()
}

// FFmpeg - video cutter
func cutVideo(inputFile, startTime, endTime, outputFile string) error {
	cmd := exec.Command("ffmpeg", "-i", tmpPath+inputFile, "-ss", startTime, "-to", endTime, "-c", "copy", clipsPath+outputFile)
	if err := cmd.Run(); err != nil {
		return err
	}

	//Clean temp dir
	if err := cleanTempDir(); err != nil {
		return err
	}

	//Create thumbnail
	if createThumbnail {
		thumbCmd := exec.Command("ffmpeg", "-i", clipsPath+outputFile, "-ss", "00:00:02", "-vf", "scale=640:360", "-vframes", "1", clipsPath+filepath.Base(outputFile)+"_thumb.jpg")
		err := thumbCmd.Run()
		if err != nil {
			return fmt.Errorf("| error | ffmpeg | thumbnail could not be created | %v", err)
		}
	}

	return nil
}
