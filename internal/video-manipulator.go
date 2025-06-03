package internal

import "sort"

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
	// Detectar se é uma borda
	if p.isEdgePixel() {
		return // Não processa bordas
	}

	// Coletar valores dos pixels vizinhos
	var values []uint8
	for y, row := range p.Pixels {
		for x, pixelValue := range row {
			if y == p.CenterY && x == p.CenterX {
				continue
			}
			values = append(values, pixelValue)
		}
	}

	if len(values) == 0 {
		return
	}

	// Calcular variância para detectar movimento
	variance := p.calculateVariance(values)
	current := p.Pixels[p.CenterY][p.CenterX]

	// Aplicar filtro baseado na variância
	if variance < 25 { // Região estável
		sort.Slice(values, func(i, j int) bool {
			return values[i] < values[j]
		})
		median := values[len(values)/2]

		alpha := 0.2 // Filtro muito conservador
		p.Pixels[p.CenterY][p.CenterX] = uint8(float64(current)*(1-alpha) + float64(median)*alpha)
	}
	// Região com movimento: preserva pixel original
}

func (p PixelsRadius) isEdgePixel() bool {
	if len(p.Pixels) < 3 || len(p.Pixels[0]) < 3 {
		return true
	}

	center := float64(p.Pixels[p.CenterY][p.CenterX])

	// Calcular gradiente com vizinhos diretos
	var gradient float64
	neighbors := []float64{}

	// Vizinhos diretos (cruz)
	if p.CenterY > 0 {
		neighbors = append(neighbors, float64(p.Pixels[p.CenterY-1][p.CenterX]))
	}
	if p.CenterY < len(p.Pixels)-1 {
		neighbors = append(neighbors, float64(p.Pixels[p.CenterY+1][p.CenterX]))
	}
	if p.CenterX > 0 {
		neighbors = append(neighbors, float64(p.Pixels[p.CenterY][p.CenterX-1]))
	}
	if p.CenterX < len(p.Pixels[0])-1 {
		neighbors = append(neighbors, float64(p.Pixels[p.CenterY][p.CenterX+1]))
	}

	// Calcular gradiente máximo
	for _, neighbor := range neighbors {
		diff := center - neighbor
		if diff < 0 {
			diff = -diff
		}
		if diff > gradient {
			gradient = diff
		}
	}

	return gradient > 15 // Threshold para bordas
}

func (p PixelsRadius) calculateVariance(values []uint8) float64 {
	if len(values) <= 1 {
		return 0
	}

	var sum, sumSq float64
	for _, v := range values {
		val := float64(v)
		sum += val
		sumSq += val * val
	}

	mean := sum / float64(len(values))
	return (sumSq / float64(len(values))) - (mean * mean)
}
