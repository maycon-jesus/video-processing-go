package internal

import "math"

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

func (p PixelsRadius) IsEdgePixel(threshold float64) bool {
	if len(p.Pixels) < 3 || len(p.Pixels[0]) < 3 {
		return true // Muito pequeno para calcular gradiente
	}

	centerY, centerX := p.CenterY, p.CenterX

	// Verificar se temos pixels suficientes ao redor do centro
	if centerY == 0 || centerY >= len(p.Pixels)-1 ||
		centerX == 0 || centerX >= len(p.Pixels[0])-1 {
		return true // Na borda da região
	}

	// Calcular gradiente usando diferenças centrais
	top := float64(p.Pixels[centerY-1][centerX])
	bottom := float64(p.Pixels[centerY+1][centerX])
	left := float64(p.Pixels[centerY][centerX-1])
	right := float64(p.Pixels[centerY][centerX+1])

	gradientX := math.Abs(right-left) / 2.0
	gradientY := math.Abs(bottom-top) / 2.0
	gradient := math.Sqrt(gradientX*gradientX + gradientY*gradientY)

	return gradient > threshold
}

func (p PixelsRadius) ApplyMedianMask() {
	// Verificar se é uma borda antes de aplicar o filtro
	if p.IsEdgePixel(20.0) { // threshold de 20 para detectar bordas
		return // Não aplicar filtro em bordas
	}

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

	if pixelsCount > 0 {
		p.Pixels[p.CenterY][p.CenterX] = uint8(sum / pixelsCount)
	}
}

// Versão alternativa com threshold customizável
func (p PixelsRadius) ApplyMedianMaskWithEdgeDetection(edgeThreshold float64) {
	if p.IsEdgePixel(edgeThreshold) {
		return // Preservar pixel original em bordas
	}

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

	if pixelsCount > 0 {
		p.Pixels[p.CenterY][p.CenterX] = uint8(sum / pixelsCount)
	}
}
