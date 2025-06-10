package internal

import (
	"math"
	"runtime"
	"sort"
	"sync"
)

func median(values []uint8) uint8 {
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	mid := len(values) / 2
	return values[mid]
}

func isEdgePixel(videoFrames VideoFrames, currentFrame, line, pixel int) bool {
	frame := videoFrames[currentFrame]
	if line == 0 || line >= len(frame)-1 || pixel == 0 || pixel >= len(frame[line])-1 {
		return true
	}

	// Usar operador Sobel para melhor detecção de bordas
	gx := float64(-int(frame[line-1][pixel-1]) + int(frame[line-1][pixel+1]) +
		-2*int(frame[line][pixel-1]) + 2*int(frame[line][pixel+1]) +
		-int(frame[line+1][pixel-1]) + int(frame[line+1][pixel+1]))

	gy := float64(-int(frame[line-1][pixel-1]) - 2*int(frame[line-1][pixel]) - int(frame[line-1][pixel+1]) +
		int(frame[line+1][pixel-1]) + 2*int(frame[line+1][pixel]) + int(frame[line+1][pixel+1]))

	gradient := math.Sqrt(gx*gx + gy*gy)
	return gradient > 25 // Threshold ajustado para P&B
}

// Detecta borrões (valores muito baixos anômalos)
func isBlur(values []uint8, current uint8) bool {
	if len(values) < 3 {
		return false
	}

	// Calcular mediana dos valores históricos
	sorted := make([]uint8, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	median := sorted[len(sorted)/2]

	// Se o valor atual é muito mais escuro que a mediana, pode ser borrão
	diff := int(median) - int(current)
	return diff > 40 && current < 60 // Valor muito escuro comparado ao histórico
}

// Detecta clarões (valores muito altos anômalos)
func isFlare(values []uint8, current uint8) bool {
	if len(values) < 3 {
		return false
	}

	// Calcular mediana dos valores históricos
	sorted := make([]uint8, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	median := sorted[len(sorted)/2]

	// Se o valor atual é muito mais claro que a mediana, pode ser clarão
	diff := int(current) - int(median)
	return diff > 50 && current > 180 // Valor muito claro comparado ao histórico
}

func isNoise(values []uint8, current uint8, variance float64) bool {
	if len(values) < 3 {
		return false
	}

	// Contar quantos valores históricos são similares entre si
	similarCount := 0
	threshold := uint8(5) // Aumentado para melhor detecção

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

	// Se a maioria dos valores históricos são similares,
	// mas o atual é muito diferente, provavelmente é ruído
	totalPairs := len(values) * (len(values) - 1) / 2
	stabilityRatio := float64(similarCount) / float64(totalPairs)

	if stabilityRatio > 0.6 { // 60% dos valores históricos são estáveis
		// Verificar se o valor atual destoa
		sort.Slice(values, func(i, j int) bool {
			return values[i] < values[j]
		})
		median := values[len(values)/2]

		currentDiff := int(current) - int(median)
		if currentDiff < 0 {
			currentDiff = -currentDiff
		}

		return uint8(currentDiff) > 12 // Aumentado para melhor detecção
	}

	return false
}

// Verifica se a região tem movimento significativo
func hasMovement(values []uint8) bool {
	if len(values) < 3 {
		return false
	}

	variance := calculateVariance(values)
	return variance > 30 // Threshold para detectar movimento
}

// Filtro adaptativo baseado na análise temporal
func adaptiveTemporalFilter(values []uint8, current uint8, variance float64) uint8 {
	if len(values) == 0 {
		return current
	}

	sorted := make([]uint8, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	median := sorted[len(sorted)/2]

	// Escolher alpha baseado na estabilidade da região
	var alpha float64
	if variance < 10 {
		alpha = 0.6 // Região muito estável - filtro mais forte
	} else if variance < 25 {
		alpha = 0.4 // Região moderadamente estável
	} else {
		alpha = 0.2 // Região com movimento - filtro suave
	}

	result := alpha*float64(median) + (1-alpha)*float64(current)

	if result < 0 {
		return 0
	}
	if result > 255 {
		return 255
	}

	return uint8(result)
}

func TimeTravalerProcessLine(videoFrames VideoFrames, currentFrame int, previousFrames int, line int) []uint8 {
	if currentFrame <= 2 {
		return videoFrames[currentFrame][line]
	}

	lineWidth := len(videoFrames[currentFrame][line])
	nLine := make([]uint8, lineWidth)
	tempValues := make([]uint8, previousFrames)
	frameStart := currentFrame - previousFrames

	for i := 0; i < lineWidth; i++ {
		current := videoFrames[currentFrame][line][i]

		// Se for uma borda, preserve o pixel original
		if isEdgePixel(videoFrames, currentFrame, line, i) {
			nLine[i] = current
			continue
		}

		// Coletar valores históricos
		for j := 0; j < previousFrames; j++ {
			tempValues[j] = videoFrames[frameStart+j][line][i]
		}

		variance := calculateVariance(tempValues)

		// Detectar e corrigir borrões
		if isBlur(tempValues, current) {
			// Usar mediana dos valores históricos mais estáveis
			sorted := make([]uint8, len(tempValues))
			copy(sorted, tempValues)
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i] < sorted[j]
			})

			// Usar um valor ligeiramente acima da mediana para corrigir borrão
			medianIdx := len(sorted) / 2
			correctedValue := sorted[medianIdx]
			if medianIdx < len(sorted)-1 {
				correctedValue = uint8((int(sorted[medianIdx]) + int(sorted[medianIdx+1])) / 2)
			}

			alpha := 0.8 // Correção agressiva para borrões
			nLine[i] = uint8(alpha*float64(correctedValue) + (1-alpha)*float64(current))

		} else if isFlare(tempValues, current) {
			// Corrigir clarões usando valores históricos mais escuros
			sorted := make([]uint8, len(tempValues))
			copy(sorted, tempValues)
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i] < sorted[j]
			})

			// Usar um valor ligeiramente abaixo da mediana para corrigir clarão
			medianIdx := len(sorted) / 2
			correctedValue := sorted[medianIdx]
			if medianIdx > 0 {
				correctedValue = uint8((int(sorted[medianIdx]) + int(sorted[medianIdx-1])) / 2)
			}

			alpha := 0.8 // Correção agressiva para clarões
			nLine[i] = uint8(alpha*float64(correctedValue) + (1-alpha)*float64(current))

		} else if isNoise(tempValues, current, variance) {
			// Ruído detectado - usar mediana
			median := median(tempValues)
			alpha := 0.7
			nLine[i] = uint8(alpha*float64(median) + (1-alpha)*float64(current))

		} else if variance < 20 && !hasMovement(tempValues) {
			// Região estável - aplicar filtro temporal adaptativo
			nLine[i] = adaptiveTemporalFilter(tempValues, current, variance)

		} else {
			// Região com movimento significativo - preservar
			nLine[i] = current
		}
	}

	return nLine
}

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
	return (sumSq / float64(len(values))) - (mean * mean)
}

func TimeTravaler(videoFrames VideoFrames, currentFrame int, previousFrames int) {
	if currentFrame <= previousFrames-1 {
		return
	}

	frame := videoFrames[currentFrame]
	totalLines := len(frame)

	// Use número de CPUs disponíveis
	numWorkers := runtime.NumCPU()

	var wg sync.WaitGroup
	lineChan := make(chan int, totalLines)

	// Envia índices das linhas
	go func() {
		for i := 0; i < totalLines; i++ {
			lineChan <- i
		}
		close(lineChan)
	}()

	// Workers processam linhas
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for lineIdx := range lineChan {
				processedLine := TimeTravalerProcessLine(videoFrames, currentFrame, previousFrames, lineIdx)
				frame[lineIdx] = processedLine
			}
		}()
	}

	wg.Wait()
}
