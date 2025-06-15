package internal

import (
	"math"
	"sort"
)

// Frame representa um único quadro em um vídeo, como uma fatia 2D de uint8 (pixels).
type Frame = [][]uint8

// VideoFrames representa uma coleção de quadros, essencialmente um vídeo.
type VideoFrames = []Frame

// FrameIndentifier é uma estrutura para armazenar um quadro e seu identificador.
type FrameIndentifier struct {
	Id     int
	Pixels Frame
}

// PixelsRadius representa uma região circular de pixels em torno de um ponto central.
// Armazena as coordenadas do centro, os dados dos pixels e a caixa delimitadora da região.
type PixelsRadius struct {
	CenterX                int   // Coordenada X do pixel central dentro da fatia Pixels.
	CenterY                int   // Coordenada Y do pixel central dentro da fatia Pixels.
	OriginalY              int   // Coordenada Y original do pixel central no quadro completo.
	OriginalX              int   // Coordenada X original do pixel central no quadro completo.
	Pixels                 Frame // Os dados reais dos pixels do raio.
	XMin, XMax, YMin, YMax int   // Coordenadas da caixa delimitadora do raio no quadro completo.
}

// GetPixelRadius extrai uma região quadrada de pixels (um "raio") de um determinado quadro.
// Recebe o quadro, as coordenadas centrais (y, x) e o tamanho do raio.
// Retorna uma estrutura PixelsRadius contendo os pixels extraídos e informações relacionadas.
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

	// Calcula as dimensões da região de pixels.
	regionHeight := yMax - yMin + 1
	regionWidth := xMax - xMin + 1
	// Cria uma cópia dos dados dos pixels dentro do raio definido.
	pixelsCopy := make([][]uint8, regionHeight)

	for i := 0; i < regionHeight; i++ {
		pixelsCopy[i] = make([]uint8, regionWidth)
		// Copia a parte relevante do quadro original para o novo pixelsCopy.
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

// IsEdgePixel determina se o pixel central do PixelsRadius é um pixel de borda.
// Utiliza um cálculo de gradiente simples (semelhante ao Sobel) e um limiar.
// Retorna verdadeiro se o gradiente estiver acima do limiar, indicando uma borda.
func (p PixelsRadius) IsEdgePixel(threshold float64) bool {
	// Se a região de pixels for muito pequena para ter vizinhos, considera-se uma borda.
	if len(p.Pixels) < 3 || len(p.Pixels[0]) < 3 {
		return true
	}

	centerY, centerX := p.CenterY, p.CenterX

	// Se o pixel central estiver na borda da fatia Pixels, considera-se uma borda.
	if centerY == 0 || centerY >= len(p.Pixels)-1 ||
		centerX == 0 || centerX >= len(p.Pixels[0])-1 {
		return true
	}

	// Obtém os valores dos pixels vizinhos.
	top := float64(p.Pixels[centerY-1][centerX])
	bottom := float64(p.Pixels[centerY+1][centerX])
	left := float64(p.Pixels[centerY][centerX-1])
	right := float64(p.Pixels[centerY][centerX+1])

	// Calcula o gradiente nas direções X e Y.
	gradientX := math.Abs(right-left) / 2.0
	gradientY := math.Abs(bottom-top) / 2.0
	// Calcula a magnitude do gradiente.
	gradient := math.Sqrt(gradientX*gradientX + gradientY*gradientY)

	// Compara a magnitude do gradiente com o limiar.
	return gradient > threshold
}

// CalculateVariance calcula a variância dos valores dos pixels dentro do PixelsRadius.
// A variância é uma medida de quão dispersos estão os valores dos pixels.
func (p PixelsRadius) CalculateVariance() float64 {
	if len(p.Pixels) == 0 {
		return 0
	}

	var sum, count float64
	// Calcula a soma de todos os valores de pixels e a contagem total de pixels.
	for _, row := range p.Pixels {
		for _, pixel := range row {
			sum += float64(pixel)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	// Calcula o valor médio (média) do pixel.
	mean := sum / count
	var variance float64

	// Calcula a soma das diferenças quadradas da média.
	for _, row := range p.Pixels {
		for _, pixel := range row {
			diff := float64(pixel) - mean
			variance += diff * diff
		}
	}

	// Retorna a variância (média das diferenças quadradas).
	return variance / count
}

// IsNoisePixel determina se o pixel central do PixelsRadius é provavelmente um pixel de ruído.
// Verifica a similaridade do pixel central com seus vizinhos.
// Se a razão de similaridade estiver abaixo de um limiar, é considerado ruído.
func (p PixelsRadius) IsNoisePixel() bool {
	centerPixel := p.Pixels[p.CenterY][p.CenterX]
	centerValue := float64(centerPixel)

	similar := 0      // Contagem de vizinhos semelhantes ao pixel central.
	total := 0        // Número total de vizinhos.
	threshold := 15.0 // Limiar para determinar similaridade.

	// Itera sobre todos os pixels no raio.
	for y, row := range p.Pixels {
		for x, pixel := range row {
			// Pula o próprio pixel central.
			if y == p.CenterY && x == p.CenterX {
				continue
			}
			total++
			// Se a diferença absoluta entre o vizinho e o centro estiver dentro do limiar, conta como similar.
			if math.Abs(float64(pixel)-centerValue) <= threshold {
				similar++
			}
		}
	}

	if total == 0 {
		// Nenhum vizinho para comparar.
		return false
	}

	// Calcula a razão de vizinhos similares para o total de vizinhos.
	similarityRatio := float64(similar) / float64(total)
	// Se a razão for baixa, o pixel é considerado ruído.
	return similarityRatio < 0.3
}

// ApplyAdaptiveFilter aplica um filtro ao pixel central do PixelsRadius.
// O tipo de filtro aplicado depende se o pixel é uma borda, ruído ou com base na variância.
func (p PixelsRadius) ApplyAdaptiveFilter() {
	if len(p.Pixels) == 0 {
		return
	}

	// Calcula as propriedades da região do pixel.
	variance := p.CalculateVariance()
	isEdge := p.IsEdgePixel(25.0) // Limiar para detecção de borda.
	isNoise := p.IsNoisePixel()

	centerPixel := p.Pixels[p.CenterY][p.CenterX]

	// Coleta todos os pixels vizinhos.
	var neighbors []uint8
	for y, row := range p.Pixels {
		for x, pixel := range row {
			if y != p.CenterY || x != p.CenterX { // Exclui o pixel central.
				neighbors = append(neighbors, pixel)
			}
		}
	}

	if len(neighbors) == 0 {
		// Sem vizinhos, nada para filtrar.
		return
	}

	// Aplica diferentes filtros com base nas características do pixel.
	switch {
	case isEdge:
		// Para pixels de borda, aplica um filtro suave com um alfa pequeno para preservar as bordas.
		p.Pixels[p.CenterY][p.CenterX] = p.applySoftFilter(neighbors, centerPixel, 0.1)

	case isNoise:
		// Para pixels de ruído, aplica um filtro de mediana para remover o ruído.
		p.Pixels[p.CenterY][p.CenterX] = p.applyMedianFilter(neighbors)

	case variance < 50:
		// Para regiões de baixa variância (áreas suaves), aplica um filtro de média com um alfa maior para suavização mais forte.
		p.Pixels[p.CenterY][p.CenterX] = p.applyMeanFilter(neighbors, centerPixel, 0.7)

	case variance < 200:
		// Para regiões de média variância, aplica um filtro suave com um alfa moderado.
		p.Pixels[p.CenterY][p.CenterX] = p.applySoftFilter(neighbors, centerPixel, 0.3)

	default:
		// Para regiões de alta variância (áreas texturizadas), aplica um filtro suave com um alfa muito pequeno para preservar os detalhes.
		p.Pixels[p.CenterY][p.CenterX] = p.applySoftFilter(neighbors, centerPixel, 0.05)
	}
}

// applyMedianFilter substitui o pixel central pelo valor mediano de seus vizinhos.
// Isso é eficaz para remover ruído do tipo sal e pimenta.
func (p PixelsRadius) applyMedianFilter(neighbors []uint8) uint8 {
	if len(neighbors) == 0 {
		return p.Pixels[p.CenterY][p.CenterX] // Retorna o pixel atual se não houver vizinhos.
	}

	// Cria uma cópia dos vizinhos para ordenar.
	sorted := make([]uint8, len(neighbors))
	copy(sorted, neighbors)
	// Ordena os vizinhos.
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Retorna o valor mediano.
	return sorted[len(sorted)/2]
}

// applyMeanFilter substitui o pixel central por uma média ponderada de si mesmo e da média de seus vizinhos.
// Alpha controla o peso da média dos vizinhos.
func (p PixelsRadius) applyMeanFilter(neighbors []uint8, current uint8, alpha float64) uint8 {
	if len(neighbors) == 0 {
		return current // Retorna o pixel atual se não houver vizinhos.
	}

	var sum int
	// Calcula a soma dos valores dos pixels vizinhos.
	for _, pixel := range neighbors {
		sum += int(pixel)
	}

	// Calcula a média dos valores dos pixels vizinhos.
	mean := float64(sum) / float64(len(neighbors))
	// Calcula o novo valor do pixel como uma média ponderada.
	result := alpha*mean + (1-alpha)*float64(current)

	// Limita o resultado ao intervalo de pixel válido [0, 255].
	if result < 0 {
		return 0
	}
	if result > 255 {
		return 255
	}

	return uint8(result)
}

// applySoftFilter substitui o pixel central por uma média ponderada de si mesmo e da mediana de seus vizinhos.
// Alpha controla o peso da mediana dos vizinhos. Este é um filtro mais suave que o filtro de média.
func (p PixelsRadius) applySoftFilter(neighbors []uint8, current uint8, alpha float64) uint8 {
	if len(neighbors) == 0 {
		return current // Retorna o pixel atual se não houver vizinhos.
	}

	// Cria uma cópia dos vizinhos para ordenar.
	sorted := make([]uint8, len(neighbors))
	copy(sorted, neighbors)
	// Ordena os vizinhos.
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Obtém a mediana dos vizinhos.
	median := sorted[len(sorted)/2]
	// Calcula o novo valor do pixel como uma média ponderada.
	result := alpha*float64(median) + (1-alpha)*float64(current)

	// Limita o resultado ao intervalo de pixel válido [0, 255].
	if result < 0 {
		return 0
	}
	if result > 255 {
		return 255
	}

	return uint8(result)
}
