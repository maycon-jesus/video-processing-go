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

	// Calcular gradiente (diferença com vizinhos)
	top := float64(frame[line-1][pixel])
	bottom := float64(frame[line+1][pixel])
	left := float64(frame[line][pixel-1])
	right := float64(frame[line][pixel+1])

	gradientX := math.Abs(right-left) / 2.0
	gradientY := math.Abs(bottom-top) / 2.0
	gradient := math.Sqrt(gradientX*gradientX + gradientY*gradientY)

	return gradient > 20 // Threshold para detectar bordas
}

func isNoise(values []uint8, current uint8, variance float64) bool {
	if len(values) < 3 {
		return false
	}

	// Contar quantos valores históricos são similares entre si
	similarCount := 0
	threshold := uint8(3)

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

		return uint8(currentDiff) > 8 // Valor atual muito diferente da mediana
	}

	return false
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

		for j := 0; j < previousFrames; j++ {
			tempValues[j] = videoFrames[frameStart+j][line][i]
		}

		variance := calculateVariance(tempValues)

		// Detector de ruído mais sofisticado
		if isNoise(tempValues, current, variance) {
			// É ruído - usar mediana dos valores históricos
			sort.Slice(tempValues, func(i, j int) bool {
				return tempValues[i] < tempValues[j]
			})
			median := tempValues[len(tempValues)/2]

			// Correção suave do ruído
			alpha := 0.7 // Mais agressivo para ruído
			nLine[i] = uint8(alpha*float64(median) + (1-alpha)*float64(current))
		} else if variance < 15 { // Região muito estável
			median := median(tempValues)
			alpha := 0.3 // Filtro moderado
			nLine[i] = uint8(alpha*float64(median) + (1-alpha)*float64(current))
		} else {
			// Região com movimento - preservar
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
