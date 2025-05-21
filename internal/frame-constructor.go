package internal

func ApplyChanges(frame Frame, radius []PixelsRadius) {
	for _, pixels := range radius {
		frame[pixels.OriginalY][pixels.OriginalX] = pixels.Pixels[pixels.CenterY][pixels.CenterX]
	}
}
