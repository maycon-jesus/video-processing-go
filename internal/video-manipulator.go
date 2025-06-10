package internal

import (
	"math"
	"sort"
)

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

	// Copiar apenas a região necessária
	regionHeight := yMax - yMin + 1
	regionWidth := xMax - xMin + 1
	pixelsCopy := make([][]uint8, regionHeight)

	for i := 0; i < regionHeight; i++ {
		pixelsCopy[i] = make([]uint8, regionWidth)
		copy(pixelsCopy[i], frame[yMin+i][xMin:xMax+1])
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
		return true
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

func (p PixelsRadius) CalculateVariance() float64 {
	if len(p.Pixels) == 0 {
		return 0
	}

	var sum, count float64
	for _, row := range p.Pixels {
		for _, pixel := range row {
			sum += float64(pixel)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	mean := sum / count
	var variance float64

	for _, row := range p.Pixels {
		for _, pixel := range row {
			diff := float64(pixel) - mean
			variance += diff * diff
		}
	}

	return variance / count
}

func (p PixelsRadius) IsNoisePixel() bool {
	centerPixel := p.Pixels[p.CenterY][p.CenterX]
	centerValue := float64(centerPixel)

	// Contar pixels similares na vizinhança
	similar := 0
	total := 0
	threshold := 15.0

	for y, row := range p.Pixels {
		for x, pixel := range row {
			if y == p.CenterY && x == p.CenterX {
				continue
			}
			total++
			if math.Abs(float64(pixel)-centerValue) <= threshold {
				similar++
			}
		}
	}

	if total == 0 {
		return false
	}

	// Se poucos pixels são similares, pode ser ruído
	similarityRatio := float64(similar) / float64(total)
	return similarityRatio < 0.3 // Menos de 30% similares = possível ruído
}

// Filtro adaptativo para reduzir granulações
func (p PixelsRadius) ApplyAdaptiveFilter() {
	if len(p.Pixels) == 0 {
		return
	}

	variance := p.CalculateVariance()
	isEdge := p.IsEdgePixel(25.0) // Threshold mais baixo para P&B
	isNoise := p.IsNoisePixel()

	centerPixel := p.Pixels[p.CenterY][p.CenterX]

	// Coletar todos os pixels da vizinhança (exceto o centro)
	var neighbors []uint8
	for y, row := range p.Pixels {
		for x, pixel := range row {
			if y != p.CenterY || x != p.CenterX {
				neighbors = append(neighbors, pixel)
			}
		}
	}

	if len(neighbors) == 0 {
		return
	}

	// Escolher filtro baseado nas características da região
	switch {
	case isEdge:
		// Preservar bordas - filtro mínimo
		p.Pixels[p.CenterY][p.CenterX] = p.applySoftFilter(neighbors, centerPixel, 0.1)

	case isNoise:
		// Ruído detectado - filtro mais agressivo
		p.Pixels[p.CenterY][p.CenterX] = p.applyMedianFilter(neighbors)

	case variance < 50: // Região muito homogênea
		// Aplicar filtro de média para suavizar granulações
		p.Pixels[p.CenterY][p.CenterX] = p.applyMeanFilter(neighbors, centerPixel, 0.7)

	case variance < 200: // Região moderadamente homogênea
		// Filtro suave
		p.Pixels[p.CenterY][p.CenterX] = p.applySoftFilter(neighbors, centerPixel, 0.3)

	default:
		// Região com muito detalhe - preservar
		// Aplicar apenas filtro muito suave
		p.Pixels[p.CenterY][p.CenterX] = p.applySoftFilter(neighbors, centerPixel, 0.05)
	}
}

func (p PixelsRadius) applyMedianFilter(neighbors []uint8) uint8 {
	if len(neighbors) == 0 {
		return p.Pixels[p.CenterY][p.CenterX]
	}

	sorted := make([]uint8, len(neighbors))
	copy(sorted, neighbors)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	return sorted[len(sorted)/2]
}

func (p PixelsRadius) applyMeanFilter(neighbors []uint8, current uint8, alpha float64) uint8 {
	if len(neighbors) == 0 {
		return current
	}

	var sum int
	for _, pixel := range neighbors {
		sum += int(pixel)
	}

	mean := float64(sum) / float64(len(neighbors))
	result := alpha*mean + (1-alpha)*float64(current)

	if result < 0 {
		return 0
	}
	if result > 255 {
		return 255
	}

	return uint8(result)
}

func (p PixelsRadius) applySoftFilter(neighbors []uint8, current uint8, alpha float64) uint8 {
	if len(neighbors) == 0 {
		return current
	}

	// Usar mediana para ser mais resistente a ruídos
	sorted := make([]uint8, len(neighbors))
	copy(sorted, neighbors)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	median := sorted[len(sorted)/2]
	result := alpha*float64(median) + (1-alpha)*float64(current)

	if result < 0 {
		return 0
	}
	if result > 255 {
		return 255
	}

	return uint8(result)
}

// Função de conveniência para aplicar filtro com configuração otimizada para P&B
func (p PixelsRadius) ApplyDenoiseFilter() {
	p.ApplyAdaptiveFilter()
}

// Versão legada mantida para compatibilidade
func (p PixelsRadius) ApplyMedianMask() {
	if p.IsEdgePixel(20.0) {
		return
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
