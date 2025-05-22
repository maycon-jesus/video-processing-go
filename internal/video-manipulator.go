package internal

type Frame = [][]uint8
type VideoFrames = []Frame

type FrameIndentifier struct {
	Id     int
	Pixels Frame
}

type PixelsRadius struct {
	CenterX                int
	CenterY                int
	OriginalY              int
	OriginalX              int
	Pixels                 Frame
	XMin, XMax, YMin, YMax int
}

func GetPixelRadius(frame Frame, y int, x int, radius int) PixelsRadius {
	yMin := y - radius
	yMax := y + radius

	if yMin < 0 {
		yMin = 0
	}
	if frameSizeY := len(frame); yMax >= frameSizeY {
		yMax = frameSizeY - 1
	}

	xMin := x - radius
	xMax := x + radius
	if xMin < 0 {
		xMin = 0
	}
	if frameSizeX := len(frame[0]); xMax >= frameSizeX {
		xMax = frameSizeX - 1
	}

	pixelsCopy := [][]uint8{}

	pixelsCopy = append(pixelsCopy, frame[yMin:yMax+1]...)
	for i, row := range pixelsCopy {
		pixelsCopy[i] = row[xMin : xMax+1]
	}

	return PixelsRadius{
		CenterX:   x - xMin,
		CenterY:   y - yMin,
		OriginalX: x,
		OriginalY: y,
		Pixels:    pixelsCopy,
		YMin:      yMin,
		YMax:      yMax,
		XMin:      xMin,
		XMax:      xMax,
	}
}

func (p PixelsRadius) ApplyMedianMask() {
	sum := 0
	pixelsCount := 0
	for y, row := range p.Pixels {
		for x, pixelValue := range row {
			if y == p.CenterY && x == p.CenterX {
				continue
			}
			sum += int(pixelValue)
			pixelsCount++
		}
	}
	p.Pixels[p.CenterY][p.CenterX] = uint8(sum / pixelsCount)
}
