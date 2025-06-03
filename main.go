package main

import (
	"fmt"
	"video-processor/internal"

	"gocv.io/x/gocv"
)

func carregarVideo(caminho string) [][][]uint8 {
	captura, err := gocv.VideoCaptureFile(caminho)
	if err != nil || !captura.IsOpened() {
		fmt.Println("Vídeo está sendo processado por outra aplicação")
		return nil
	}
	defer captura.Close()

	// Tamanho do frame
	largura := int(captura.Get(gocv.VideoCaptureFrameWidth))
	altura := int(captura.Get(gocv.VideoCaptureFrameHeight))

	var frames [][][]uint8
	matRGB := gocv.NewMat()
	matCinza := gocv.NewMatWithSize(altura, largura, gocv.MatTypeCV8U)
	defer matRGB.Close()
	defer matCinza.Close()

	for {
		if ok := captura.Read(&matRGB); !ok || matRGB.Empty() {
			break
		}

		// Converter para escala de cinza
		gocv.CvtColor(matRGB, &matCinza, gocv.ColorBGRToGray)

		// Copiar os pixels
		pixels := make([][]uint8, altura)
		for y := 0; y < altura; y++ {
			row := matCinza.RowRange(y, y+1)
			pixels[y] = make([]uint8, largura)
			copy(pixels[y], row.ToBytes())
			row.Close()
		}
		frames = append(frames, pixels)
	}

	fmt.Printf("%d x %d, %d frames\n", largura, altura, len(frames))
	return frames
}

func gravarVideo(frames [][][]uint8, caminho string, fps float64) {
	if len(frames) == 0 {
		fmt.Println("Nenhum frame para gravar")
		return
	}

	altura := len(frames[0])
	largura := len(frames[0][0])
	numBytes := largura * altura * 3

	writer, err := gocv.VideoWriterFile(caminho, "avc1", fps, largura, altura, true)
	if err != nil {
		fmt.Println("Erro ao abrir escritor de vídeo:", err)
		return
	}
	defer writer.Close()

	matRGB := gocv.NewMatWithSize(altura, largura, gocv.MatTypeCV8UC3)
	defer matRGB.Close()

	for _, frame := range frames {
		// Cria um slice com os dados RGB de um frame completo
		data := make([]byte, numBytes)
		idx := 0
		for y := 0; y < altura; y++ {
			for x := 0; x < largura; x++ {
				g := frame[y][x]
				data[idx] = g   // B
				data[idx+1] = g // G
				data[idx+2] = g // R
				idx += 3
			}
		}

		mat, err := gocv.NewMatFromBytes(altura, largura, gocv.MatTypeCV8UC3, data)
		if err != nil {
			fmt.Println("Erro ao criar Mat do frame:", err)
			continue
		}
		writer.Write(mat)
		mat.Close()
	}
}

func main() {
	fmt.Println("ola mundo")
	caminhoVideo := "./videos/video.mp4"
	caminhoSaida := "./videos/video2.mp4"
	caminhoSaidaOriginal := "./videos/video3.mp4"
	fps := 24.0

	fmt.Println("→ Lendo", caminhoVideo)
	pixels := carregarVideo(caminhoVideo)
	pixels = pixels[400:640]

	gravarVideo(pixels, caminhoSaidaOriginal, fps)

	if len(pixels) > 0 {
		fmt.Printf("Frames: %d   Resolução: %dx%d\n", len(pixels), len(pixels[0][0]), len(pixels[0]))
	}

	filaProcessamento := make(chan internal.FrameIndentifier)
	filaFramesProcessados := make(chan internal.FrameIndentifier, len(pixels))

	go func() {
		for i, frame := range pixels {
			filaProcessamento <- internal.FrameIndentifier{
				Id:     i,
				Pixels: frame,
			}
		}
		close(filaProcessamento)
	}()

	go func() {
		for range 22 {
			go func() {
				for {
					frame, ok := <-filaProcessamento
					if !ok {
						break
					}
					var frameCopy internal.Frame
					fmt.Println("Frame ", frame.Id)
					for y, row := range frame.Pixels {
						frameCopy = append(frameCopy, []uint8{})
						for x, _ := range row {
							radius := internal.GetPixelRadius(frame.Pixels, y, x, 4)
							radius.ApplyMedianMask()
							frameCopy[y] = append(frameCopy[y], radius.Pixels[radius.CenterY][radius.CenterX])
						}
					}
					frame.Pixels = frameCopy
					filaFramesProcessados <- frame
				}
			}()
		}
	}()

	for range len(pixels) {
		frame := <-filaFramesProcessados
		// fmt.Println("copiando frame", frame.Id)
		pixels[frame.Id] = frame.Pixels
	}
	close(filaFramesProcessados)

	for frameId := range len(pixels) {
		internal.TimeTravaler(pixels, frameId, 23)
		fmt.Println("Frame ", frameId)
	}

	//for i, frame := range pixels {
	//	var frameCopy [][]uint8
	//	fmt.Println("Frame ", i)
	//	for y, row := range frame {
	//		frameCopy = append(frameCopy, []uint8{})
	//		for x, _ := range row {
	//			radius := internal.GetPixelRadius(frame, y, x, 1)
	//			radius.ApplyMedianMask()
	//			frameCopy[y] = append(frameCopy[y], radius.Pixels[radius.CenterY][radius.CenterX])
	//		}
	//	}
	//	pixels[i] = frameCopy
	//}

	fmt.Println("→ Gravando", caminhoSaida)
	gravarVideo(pixels, caminhoSaida, fps)
	fmt.Println("Concluído!")
}
