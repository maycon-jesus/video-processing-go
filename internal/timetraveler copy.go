package internal

// import (
// 	"fmt"
// 	"math"
// 	"sort"
// )

// type TimeTravalerResponse struct {
// 	videoFrames    VideoFrames
// 	currentFrame   int
// 	previousFrames int
// 	line           int
// 	cells          []uint8
// }

// func median(values []uint8) uint8 {
// 	sort.Slice(values, func(i, j int) bool {
// 		return values[i] < values[j]
// 	})
// 	mid := len(values) / 2
// 	return values[mid]
// }

// func media(values []uint8) uint8 {
// 	sum := 0
// 	pixelsCount := 0
// 	for _, val := range values {
// 		sum += int(val)
// 		pixelsCount++
// 	}
// 	return uint8(sum / pixelsCount)
// }

// func variance(values []uint8) float64 {
// 	mean := float64(media(values))
// 	var sum float64
// 	for _, v := range values {
// 		sum += math.Pow(float64(v)-mean, 2)
// 	}
// 	return sum / float64(len(values))
// }

// func TimeTravalerProcessLine(videoFrames VideoFrames, currentFrame int, previousFrames int, line int) []uint8 {
// 	if currentFrame <= 2 {
// 		return videoFrames[currentFrame][line]
// 	}
// 	cells := make([][]uint8, len(videoFrames[currentFrame][line]))

// 	frameStart := currentFrame - previousFrames
// 	for i, frame := range videoFrames[frameStart:currentFrame] {
// 		for colI, col := range frame[line] {
// 			if currentFrame == 5 && line == 0 && colI == 0 {
// 				fmt.Println("aaa", currentFrame, line, colI)
// 			}
// 			if cells[colI] == nil {
// 				cells[colI] = make([]uint8, previousFrames)
// 				cells[colI][i] = col
// 			} else {
// 				cells[colI][i] = col
// 			}
// 		}
// 	}

// 	nLine := make([]uint8, len(videoFrames[currentFrame][line]))
// 	for i, col := range cells {
// 		// diff := math.Abs(float64(videoFrames[currentFrame][line][i]) - float64(median(col)))
// 		// diffN := int(videoFrames[currentFrame][line][i]) - int(median(col))

// 		// nLine[i] = videoFrames[currentFrame][line][i]
// 		// if diff < 5 && currentFrame-1 >= 0 {
// 		// 	nLine[i] = videoFrames[currentFrame-1][line][i]
// 		// } else if diff < 22 {
// 		// 	nLine[i] = median(col)
// 		// } else if diff > 100 {
// 		// 	nLine[i] = col[len(col)-1]
// 		// 	positive := true
// 		// 	if diffN < 0 {
// 		// 		positive = false
// 		// 	}
// 		// 	if positive {
// 		// 		nLine[i] = nLine[i] - 50
// 		// 	} else {
// 		// 		nLine[i] = nLine[i] + 50

// 		// 	}
// 		// }

// 		med := median(col)
// 		orig := videoFrames[currentFrame][line][i]
// 		diff := math.Abs(float64(orig) - float64(med))
// 		prev := videoFrames[currentFrame-1][line][i]
// 		movement := math.Abs(float64(orig) - float64(prev))

// 		if movement < 20 {
// 			if diff < 30 {
// 				alpha := diff / 30.0
// 				newVal := (1-alpha)*float64(orig) + alpha*float64(med)
// 				nLine[i] = uint8(newVal)
// 			} else {
// 				nLine[i] = med
// 			}

// 			if orig == 0 || orig == 255 {
// 				nLine[i] = med
// 			}
// 		} else {
// 			// Em vez de manter valor original puro, suavize levemente com valor anterior
// 			alpha := 0.75 // apenas 50% do valor anterior
// 			nLine[i] = uint8((1-alpha)*float64(orig) + alpha*float64(prev))
// 		}

// 	}

// 	return nLine
// }

// func TimeTravaler(videoFrames VideoFrames, currentFrame int, previousFrames int) {
// 	if currentFrame <= previousFrames-1 {
// 		return
// 	}

// 	frame := videoFrames[currentFrame]
// 	totalLines := len(frame)

// 	queueProcess := make(chan TimeTravalerResponse, totalLines)
// 	queueResponse := make(chan TimeTravalerResponse, totalLines)

// 	go func() {
// 		for j, _ := range frame {
// 			queueProcess <- TimeTravalerResponse{
// 				videoFrames:    videoFrames,
// 				currentFrame:   currentFrame,
// 				previousFrames: previousFrames,
// 				line:           j,
// 				cells:          nil,
// 			}
// 		}
// 		close(queueProcess)
// 	}()

// 	go func() {
// 		for range 22 {
// 			go func() {
// 				for {
// 					frame, ok := <-queueProcess
// 					if !ok {
// 						break
// 					}
// 					frame.cells = TimeTravalerProcessLine(frame.videoFrames, frame.currentFrame, frame.previousFrames, frame.line)
// 					queueResponse <- frame
// 				}
// 			}()
// 		}
// 	}()

// 	for range totalLines {
// 		lineProcessed := <-queueResponse
// 		// fmt.Println(frame[lineProcessed.line][0], lineProcessed.cells[0])
// 		frame[lineProcessed.line] = lineProcessed.cells
// 	}

// 	// for i := currentFrame - 1; i >= currentFrame-previousFrames; i-- {
// 	// 	actualFrame := videoFrames[i]
// 	// 	for rowI, row := range actualFrame {
// 	// 		if len(nFrame) <= rowI {
// 	// 			nFrame = append(nFrame, []uint8{})
// 	// 		}
// 	// 		for colI, col := range row {
// 	// 			if len(nFrame[rowI]) <= colI {
// 	// 				nFrame[rowI] = append(nFrame[rowI], col)
// 	// 			} else {
// 	// 				nFrame[rowI][colI] += col
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }
// }
