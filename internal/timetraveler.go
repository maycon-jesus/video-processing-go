package internal

import (
	"math"
	"runtime"
	"sort"
	"sync"
)

// median calcula a mediana de um slice de uint8.
func median(values []uint8) uint8 {
	// Cria uma cópia para não modificar o slice original.
	sortedValues := make([]uint8, len(values))
	copy(sortedValues, values)
	sort.Slice(sortedValues, func(i, j int) bool {
		return sortedValues[i] < sortedValues[j]
	})
	mid := len(sortedValues) / 2
	return sortedValues[mid]
}

// isEdgePixel verifica se um pixel é uma borda usando o operador Sobel.
func isEdgePixel(videoFrames VideoFrames, currentFrame, line, pixel int) bool {
	frame := videoFrames[currentFrame]
	// Verifica se o pixel está nas bordas do frame.
	if line == 0 || line >= len(frame)-1 || pixel == 0 || pixel >= len(frame[line])-1 {
		return true
	}

	// Operador Sobel para detecção de bordas.
	gx := float64(-int(frame[line-1][pixel-1]) + int(frame[line-1][pixel+1]) +
		-2*int(frame[line][pixel-1]) + 2*int(frame[line][pixel+1]) +
		-int(frame[line+1][pixel-1]) + int(frame[line+1][pixel+1]))

	gy := float64(-int(frame[line-1][pixel-1]) - 2*int(frame[line-1][pixel]) - int(frame[line-1][pixel+1]) +
		int(frame[line+1][pixel-1]) + 2*int(frame[line+1][pixel]) + int(frame[line+1][pixel+1]))

	gradient := math.Sqrt(gx*gx + gy*gy)
	// Define um limiar para considerar como borda.
	return gradient > 25
}

// isBlur verifica se o pixel atual está borrado em comparação com os valores anteriores.
func isBlur(values []uint8, current uint8) bool {
	if len(values) < 3 {
		return false
	}

	sorted := make([]uint8, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	median := sorted[len(sorted)/2]

	diff := int(median) - int(current)
	// Condições para identificar blur: diferença significativa e valor atual baixo.
	return diff > 40 && current < 60
}

// isFlare verifica se o pixel atual é um reflexo (flare) em comparação com os valores anteriores.
func isFlare(values []uint8, current uint8) bool {
	if len(values) < 3 {
		return false
	}

	sorted := make([]uint8, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	median := sorted[len(sorted)/2]

	diff := int(current) - int(median)
	// Condições para identificar flare: diferença significativa e valor atual alto.
	return diff > 50 && current > 180
}

// isNoise verifica se o pixel atual é ruído em comparação com os valores anteriores.
func isNoise(values []uint8, current uint8, variance float64) bool {
	if len(values) < 3 {
		return false
	}

	similarCount := 0
	threshold := uint8(5) // Limiar para considerar pixels como similares.

	// Conta pares de pixels com valores próximos.
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			diff := int(values[i]) - int(values[j])
			if diff < 0 {
				diff = -diff
			}
			if uint8(diff) <= threshold {
				similarCount++
			}
		}
	}

	totalPairs := len(values) * (len(values) - 1) / 2
	stabilityRatio := float64(similarCount) / float64(totalPairs)

	// Se a maioria dos pixels anteriores forem estáveis (similares),
	// verifica se o pixel atual destoa muito da mediana.
	if stabilityRatio > 0.6 {
		// Cria uma cópia para não modificar o slice original.
		sortedValues := make([]uint8, len(values))
		copy(sortedValues, values)
		sort.Slice(sortedValues, func(i, j int) bool {
			return sortedValues[i] < sortedValues[j]
		})
		median := sortedValues[len(sortedValues)/2]

		currentDiff := int(current) - int(median)
		if currentDiff < 0 {
			currentDiff = -currentDiff
		}

		return uint8(currentDiff) > 12 // Limiar para considerar como ruído.
	}

	return false
}

// hasMovement verifica se há movimento significativo nos valores dos pixels anteriores.
func hasMovement(values []uint8) bool {
	if len(values) < 3 {
		return false
	}

	variance := calculateVariance(values)
	// Limiar de variância para detectar movimento.
	return variance > 30
}

// adaptiveTemporalFilter aplica um filtro temporal adaptativo.
// A intensidade do filtro (alpha) depende da variância dos pixels anteriores.
func adaptiveTemporalFilter(values []uint8, current uint8, variance float64) uint8 {
	if len(values) == 0 {
		return current
	}

	// Cria uma cópia para não modificar o slice original.
	sorted := make([]uint8, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	median := sorted[len(sorted)/2]

	var alpha float64
	// Ajusta o peso (alpha) com base na variância.
	// Menor variância = maior peso para a mediana dos frames anteriores.
	if variance < 10 {
		alpha = 0.6
	} else if variance < 25 {
		alpha = 0.4
	} else {
		alpha = 0.2
	}

	result := alpha*float64(median) + (1-alpha)*float64(current)

	// Garante que o resultado esteja no intervalo [0, 255].
	if result < 0 {
		return 0
	}
	if result > 255 {
		return 255
	}

	return uint8(result)
}

// TimeTravalerProcessLine processa uma única linha de um frame de vídeo.
// Aplica diferentes técnicas de filtragem temporal baseadas na análise dos pixels.
func TimeTravalerProcessLine(videoFrames VideoFrames, currentFrame int, previousFrames int, line int) []uint8 {
	// Não processa os primeiros frames, pois não há frames anteriores suficientes.
	if currentFrame <= 2 {
		return videoFrames[currentFrame][line]
	}

	lineWidth := len(videoFrames[currentFrame][line])
	nLine := make([]uint8, lineWidth)           // Linha processada.
	tempValues := make([]uint8, previousFrames) // Valores do pixel atual nos frames anteriores.
	frameStart := currentFrame - previousFrames

	for i := 0; i < lineWidth; i++ {
		current := videoFrames[currentFrame][line][i] // Pixel atual.

		// Se for um pixel de borda, mantém o valor original.
		if isEdgePixel(videoFrames, currentFrame, line, i) {
			nLine[i] = current
			continue
		}

		// Coleta os valores do pixel atual nos frames anteriores.
		for j := 0; j < previousFrames; j++ {
			tempValues[j] = videoFrames[frameStart+j][line][i]
		}

		variance := calculateVariance(tempValues) // Calcula a variância dos pixels anteriores.

		// Aplica diferentes filtros com base nas características detectadas.
		if isBlur(tempValues, current) {
			// Correção para blur: usa a média da mediana e do próximo valor ordenado.
			sorted := make([]uint8, len(tempValues))
			copy(sorted, tempValues)
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i] < sorted[j]
			})

			medianIdx := len(sorted) / 2
			correctedValue := sorted[medianIdx]
			if medianIdx < len(sorted)-1 {
				correctedValue = uint8((int(sorted[medianIdx]) + int(sorted[medianIdx+1])) / 2)
			}

			alpha := 0.8 // Peso para a correção.
			nLine[i] = uint8(alpha*float64(correctedValue) + (1-alpha)*float64(current))

		} else if isNoise(tempValues, current, variance) {
			// Correção para ruído: usa a mediana dos frames anteriores.
			medianVal := median(tempValues)
			alpha := 0.7 // Peso para a correção.
			nLine[i] = uint8(alpha*float64(medianVal) + (1-alpha)*float64(current))

		} else if variance < 20 && !hasMovement(tempValues) {
			// Se há baixa variância e pouco movimento, aplica filtro temporal adaptativo.
			nLine[i] = adaptiveTemporalFilter(tempValues, current, variance)

		} else {
			// Caso contrário, mantém o pixel original.
			nLine[i] = current
		}
	}

	return nLine
}

// calculateVariance calcula a variância de um slice de uint8.
func calculateVariance(values []uint8) float64 {
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
	// Fórmula da variância: E[X^2] - (E[X])^2
	return (sumSq / float64(len(values))) - (mean * mean)
}

// TimeTravaler processa um frame de vídeo completo, aplicando o filtro temporal em paralelo por linha.
func TimeTravaler(videoFrames VideoFrames, currentFrame int, previousFrames int) {
	// Não processa se não houver frames anteriores suficientes.
	if currentFrame <= previousFrames-1 {
		return
	}

	frame := videoFrames[currentFrame]
	totalLines := len(frame)

	numWorkers := runtime.NumCPU() // Usa o número de CPUs disponíveis como workers.

	var wg sync.WaitGroup
	lineChan := make(chan int, totalLines) // Canal para distribuir as linhas entre os workers.

	// Goroutine para popular o canal com os índices das linhas.
	go func() {
		for i := 0; i < totalLines; i++ {
			lineChan <- i
		}
		close(lineChan)
	}()

	// Inicia os workers.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Cada worker processa linhas do canal até que o canal seja fechado.
			for lineIdx := range lineChan {
				processedLine := TimeTravalerProcessLine(videoFrames, currentFrame, previousFrames, lineIdx)
				frame[lineIdx] = processedLine // Atualiza a linha no frame original.
			}
		}()
	}

	wg.Wait() // Espera todos os workers terminarem.
}
