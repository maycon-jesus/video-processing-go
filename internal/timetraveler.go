package internal

import (
	"fmt"
	"math"
	"sort"
)

type TimeTravalerResponse struct {
	videoFrames    VideoFrames
	currentFrame   int
	previousFrames int
	line           int
	cells          []uint8
}

func median(values []uint8) uint8 {
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	mid := len(values) / 2
	return values[mid]
}

func TimeTravalerProcessLine(videoFrames VideoFrames, currentFrame int, previousFrames int, line int) []uint8 {
	if currentFrame <= 2 {
		return videoFrames[currentFrame][line]
	}
	cells := make([][]uint8, len(videoFrames[currentFrame][line]))

	frameStart := currentFrame - previousFrames
	for i, frame := range videoFrames[frameStart:currentFrame] {
		for colI, col := range frame[line] {
			if currentFrame == 5 && line == 0 && colI == 0 {
				fmt.Println("aaa", currentFrame, line, colI)
			}
			if cells[colI] == nil {
				cells[colI] = make([]uint8, previousFrames)
				cells[colI][i] = col
			} else {
				cells[colI][i] = col
			}
		}
	}

	nLine := make([]uint8, len(videoFrames[currentFrame][line]))
	for i, col := range cells {
		diff := math.Abs(float64(videoFrames[currentFrame][line][i]) - float64(median(col)))
		if diff < 13 || diff > 150 {
			nLine[i] = median(col)
		} else {
			nLine[i] = videoFrames[currentFrame][line][i]

		}
	}

	return nLine
}

func TimeTravaler(videoFrames VideoFrames, currentFrame int, previousFrames int) {
	if currentFrame <= previousFrames-1 {
		return
	}

	frame := videoFrames[currentFrame]
	totalLines := len(frame)

	queueProcess := make(chan TimeTravalerResponse, totalLines)
	queueResponse := make(chan TimeTravalerResponse, totalLines)

	go func() {
		for j, _ := range frame {
			queueProcess <- TimeTravalerResponse{
				videoFrames:    videoFrames,
				currentFrame:   currentFrame,
				previousFrames: previousFrames,
				line:           j,
				cells:          nil,
			}
		}
		close(queueProcess)
	}()

	go func() {
		for range 22 {
			go func() {
				for {
					frame, ok := <-queueProcess
					if !ok {
						break
					}
					frame.cells = TimeTravalerProcessLine(frame.videoFrames, frame.currentFrame, frame.previousFrames, frame.line)
					queueResponse <- frame
				}
			}()
		}
	}()

	for range totalLines {
		lineProcessed := <-queueResponse
		// fmt.Println(frame[lineProcessed.line][0], lineProcessed.cells[0])
		frame[lineProcessed.line] = lineProcessed.cells
	}

	// for i := currentFrame - 1; i >= currentFrame-previousFrames; i-- {
	// 	actualFrame := videoFrames[i]
	// 	for rowI, row := range actualFrame {
	// 		if len(nFrame) <= rowI {
	// 			nFrame = append(nFrame, []uint8{})
	// 		}
	// 		for colI, col := range row {
	// 			if len(nFrame[rowI]) <= colI {
	// 				nFrame[rowI] = append(nFrame[rowI], col)
	// 			} else {
	// 				nFrame[rowI][colI] += col
	// 			}
	// 		}
	// 	}
	// }
}
