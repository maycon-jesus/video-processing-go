package internal

import "fmt"

type TimeTravalerResponse struct {
	videoFrames    VideoFrames
	currentFrame   int
	previousFrames int
	line           int
	cells          []uint8
}

func TimeTravalerProcessLine(videoFrames VideoFrames, currentFrame int, previousFrames int, line int) []uint8 {
	cells := make([]uint8, 0)

	videoOriginalRow := videoFrames[currentFrame][line]
	for colI, col := range videoOriginalRow {
		if len(cells) <= colI {
			cells = append(cells, col)
		} else {
			cells[colI] += col
		}
	}
	return cells
}

func TimeTravaler(videoFrames VideoFrames, currentFrame int, previousFrames int) {
	nFrame := [][]uint8{}

	queueProcess := make(chan TimeTravalerResponse)
	queueResponse := make(chan TimeTravalerResponse)

	go func() {
		for i, _ := range videoFrames {
			for j, _ := range videoFrames[i] {
				queueProcess <- TimeTravalerResponse{
					videoFrames:    videoFrames,
					currentFrame:   currentFrame,
					previousFrames: previousFrames,
					line:           j,
					cells:          nil,
				}
			}
		}
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
				}
			}()
		}
	}()

	for i := currentFrame - 1; i >= currentFrame-previousFrames; i-- {
		actualFrame := videoFrames[i]
		for rowI, row := range actualFrame {
			if len(nFrame) <= rowI {
				nFrame = append(nFrame, []uint8{})
			}
			for colI, col := range row {
				if len(nFrame[rowI]) <= colI {
					nFrame[rowI] = append(nFrame[rowI], col)
				} else {
					nFrame[rowI][colI] += col
				}
			}
		}
	}
	fmt.Println("primeiro frame", nFrame[0][0])
}
