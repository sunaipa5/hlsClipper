package hlsClipper

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var (
	//Default: ./.clipTmp/
	tmpPath = "./.clipTmp/"
	//Default: ./clips/
	clipsPath = "./clips/"
	//Default: true - Automatically generate thumbnails
	createThumbnail = true
)

// Time format: 00:00:00
func CreateClip(m3u8Content io.Reader, streamUrl, startTime, endTime, filename string) error {
	if err := checkDirs(); err != nil {
		return err
	}
	if err := checkFileExist(clipsPath + filename + ".mp4"); err != nil {
		return err
	}

	startTimeFloat := parseTime(startTime)
	endTimeFloat := parseTime(endTime)

	var segments []string

	var currentSegment string
	var currentStartTime float64

	var lastSegmentDuration float64
	var lastSegmentStartTime float64

	var downloadedSegmentDuration float64

	var cutTimeStart string
	var cutTimeEnd string

	segmentCount := 0
	scanner := bufio.NewScanner(m3u8Content)
	for scanner.Scan() {

		line := scanner.Text()
		if strings.HasPrefix(line, "#EXTINF:") {
			re := regexp.MustCompile(`#EXTINF:([\d\.]+),`)
			matches := re.FindStringSubmatch(line)

			if len(matches) > 1 {
				duration, _ := strconv.ParseFloat(matches[1], 64)

				if scanner.Scan() {
					currentSegment = scanner.Text()
					segmentEndTime := currentStartTime + duration

					if isWithinRange(currentStartTime, segmentEndTime, startTimeFloat, endTimeFloat) {
						if segmentCount == 0 {
							cutTimeStart = strconv.FormatFloat(parseTime(startTime)-currentStartTime, 'f', 2, 64)
							segmentCount = 1
						}

						if err := downloadSegment(streamUrl, currentSegment); err != nil {
							return fmt.Errorf("| error | http.Get | segment %s could not be downloaded: %v", currentSegment, err)
						}

						parsedSegmentName, err := url.Parse(currentSegment)
						if err != nil {
							return err
						}
						segments = append(segments, path.Base(parsedSegmentName.Path))

						downloadedSegmentDuration += duration
						lastSegmentDuration = duration
						lastSegmentStartTime = currentStartTime
					}

					if currentStartTime >= parseTime(endTime) {
						cutTimeEnd = strconv.FormatFloat(downloadedSegmentDuration-((lastSegmentStartTime+lastSegmentDuration)-parseTime(endTime)), 'f', 2, 64)
						break
					}
				}

				if currentSegment != "" {
					currentStartTime += duration
				} else {
					currentStartTime = 0
				}
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("| error | bufio.Scanner | scanner error: %v", err)
	}
	// Create concat file for segment merge
	concatFile := "segments.txt"
	if err := createConcatFile(segments, concatFile); err != nil {
		return fmt.Errorf("| error | os.Create | concat file could not be created: %v", err)
	}

	// Merge segments
	if err := mergeSegments(concatFile, filename+"_temp.mp4"); err != nil {
		return fmt.Errorf("| error | ffmpeg | segments could not be merged: %v", err)
	}

	// Cut Video
	if err := cutVideo(filename+"_temp.mp4", cutTimeStart, cutTimeEnd, filename+".mp4"); err != nil {
		return fmt.Errorf("| error | ffmpeg | video could not be cut: %v", err)
	}

	return nil
}

func isWithinRange(segmentStart, segmentEnd, startTime, endTime float64) bool {
	return (segmentEnd >= startTime && segmentStart <= endTime)
}

// 00:00:00 format to second
func parseTime(timeStr string) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return hours*3600 + minutes*60 + seconds
}

// Segment downloader
func downloadSegment(baseUrl, segmentURL string) error {
	resp, err := http.Get(baseUrl + segmentURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	//Save segment file
	parsedURL, err := url.Parse(segmentURL)
	if err != nil {
		return err
	}

	fileName := path.Base(parsedURL.Path)
	file, err := os.Create(tmpPath + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// Concat file creator
func createConcatFile(segments []string, concatFile string) error {
	file, err := os.Create(tmpPath + concatFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, segment := range segments {
		_, err := file.WriteString("file '" + segment + "'\n")
		if err != nil {
			return err
		}
	}

	return nil
}
