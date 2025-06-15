package internal

import (
	"testing"
)

// Helper function to create a test frame with specified dimensions and initial value
func createTestFrameForConstructor(height, width int, value uint8) Frame {
	frame := make(Frame, height)
	for i := range frame {
		frame[i] = make([]uint8, width)
		for j := range frame[i] {
			frame[i][j] = value
		}
	}
	return frame
}

// Helper function to create a PixelsRadius for testing
func createTestPixelsRadius(originalX, originalY, centerX, centerY int, pixels Frame) PixelsRadius {
	return PixelsRadius{
		OriginalX: originalX,
		OriginalY: originalY,
		CenterX:   centerX,
		CenterY:   centerY,
		Pixels:    pixels,
	}
}

// Helper function to create a small 3x3 pixel region
func create3x3PixelRegion(centerValue uint8) Frame {
	pixels := make(Frame, 3)
	for i := range pixels {
		pixels[i] = make([]uint8, 3)
		for j := range pixels[i] {
			if i == 1 && j == 1 { // Center pixel
				pixels[i][j] = centerValue
			} else {
				pixels[i][j] = centerValue / 2 // Different value for surrounding pixels
			}
		}
	}
	return pixels
}

func TestApplyChanges(t *testing.T) {
	tests := []struct {
		name           string
		originalFrame  Frame
		pixelsRadius   []PixelsRadius
		expectedResult Frame
		description    string
	}{
		{
			name:          "Single pixel change in center",
			originalFrame: createTestFrameForConstructor(5, 5, 100),
			pixelsRadius: []PixelsRadius{
				createTestPixelsRadius(2, 2, 1, 1, create3x3PixelRegion(200)),
			},
			expectedResult: func() Frame {
				frame := createTestFrameForConstructor(5, 5, 100)
				frame[2][2] = 200 // Center pixel should be changed
				return frame
			}(),
			description: "Should change the center pixel to the processed value",
		},
		{
			name:          "Multiple pixel changes",
			originalFrame: createTestFrameForConstructor(5, 5, 50),
			pixelsRadius: []PixelsRadius{
				createTestPixelsRadius(1, 1, 1, 1, create3x3PixelRegion(150)),
				createTestPixelsRadius(3, 3, 1, 1, create3x3PixelRegion(250)),
			},
			expectedResult: func() Frame {
				frame := createTestFrameForConstructor(5, 5, 50)
				frame[1][1] = 150
				frame[3][3] = 250
				return frame
			}(),
			description: "Should apply changes to multiple pixels",
		},
		{
			name:          "Edge pixel changes",
			originalFrame: createTestFrameForConstructor(3, 3, 75),
			pixelsRadius: []PixelsRadius{
				createTestPixelsRadius(0, 0, 0, 0, [][]uint8{{100}}),
				createTestPixelsRadius(2, 2, 0, 0, [][]uint8{{200}}),
			},
			expectedResult: func() Frame {
				frame := createTestFrameForConstructor(3, 3, 75)
				frame[0][0] = 100
				frame[2][2] = 200
				return frame
			}(),
			description: "Should handle edge pixels correctly",
		},
		{
			name:          "No changes - empty radius slice",
			originalFrame: createTestFrameForConstructor(4, 4, 128),
			pixelsRadius:  []PixelsRadius{},
			expectedResult: func() Frame {
				return createTestFrameForConstructor(4, 4, 128)
			}(),
			description: "Should leave frame unchanged when no radius data provided",
		},
		{
			name:          "Overlapping changes - last one wins",
			originalFrame: createTestFrameForConstructor(3, 3, 0),
			pixelsRadius: []PixelsRadius{
				createTestPixelsRadius(1, 1, 0, 0, [][]uint8{{100}}),
				createTestPixelsRadius(1, 1, 0, 0, [][]uint8{{255}}), // Same position, different value
			},
			expectedResult: func() Frame {
				frame := createTestFrameForConstructor(3, 3, 0)
				frame[1][1] = 255 // Last value should be applied
				return frame
			}(),
			description: "When multiple changes target same pixel, last one should win",
		},
		{
			name:          "Large frame with scattered changes",
			originalFrame: createTestFrameForConstructor(10, 10, 64),
			pixelsRadius: []PixelsRadius{
				createTestPixelsRadius(0, 0, 1, 1, create3x3PixelRegion(32)),
				createTestPixelsRadius(5, 5, 1, 1, create3x3PixelRegion(96)),
				createTestPixelsRadius(9, 9, 1, 1, create3x3PixelRegion(160)),
			},
			expectedResult: func() Frame {
				frame := createTestFrameForConstructor(10, 10, 64)
				frame[0][0] = 32
				frame[5][5] = 96
				frame[9][9] = 160
				return frame
			}(),
			description: "Should handle changes scattered across a larger frame",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the original frame to avoid modifying the test data
			frameCopy := make(Frame, len(tt.originalFrame))
			for i := range tt.originalFrame {
				frameCopy[i] = make([]uint8, len(tt.originalFrame[i]))
				copy(frameCopy[i], tt.originalFrame[i])
			}

			// Apply the changes
			ApplyChanges(frameCopy, tt.pixelsRadius)

			// Verify the result
			if len(frameCopy) != len(tt.expectedResult) {
				t.Errorf("Frame height mismatch: got %d, want %d", len(frameCopy), len(tt.expectedResult))
				return
			}

			for y := range frameCopy {
				if len(frameCopy[y]) != len(tt.expectedResult[y]) {
					t.Errorf("Frame width mismatch at row %d: got %d, want %d", y, len(frameCopy[y]), len(tt.expectedResult[y]))
					return
				}

				for x := range frameCopy[y] {
					if frameCopy[y][x] != tt.expectedResult[y][x] {
						t.Errorf("Pixel mismatch at (%d,%d): got %d, want %d", y, x, frameCopy[y][x], tt.expectedResult[y][x])
					}
				}
			}
		})
	}
}

func TestApplyChanges_EdgeCases(t *testing.T) {
	t.Run("Nil frame - should panic", func(t *testing.T) {
		// This test verifies that the function panics with nil input (expected behavior)
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected ApplyChanges to panic with nil frame, but it didn't")
			} else {
				t.Logf("ApplyChanges correctly panicked with nil frame: %v", r)
			}
		}()

		var frame Frame
		pixelsRadius := []PixelsRadius{
			createTestPixelsRadius(0, 0, 0, 0, [][]uint8{{100}}),
		}

		ApplyChanges(frame, pixelsRadius)
	})

	t.Run("Out of bounds coordinates - should panic", func(t *testing.T) {
		// This test verifies behavior with out-of-bounds coordinates (expected to panic)
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected ApplyChanges to panic with out-of-bounds coordinates, but it didn't")
			} else {
				t.Logf("ApplyChanges correctly panicked with out-of-bounds coordinates: %v", r)
			}
		}()

		frame := createTestFrameForConstructor(3, 3, 100)
		pixelsRadius := []PixelsRadius{
			createTestPixelsRadius(10, 10, 0, 0, [][]uint8{{200}}), // Out of bounds
		}

		ApplyChanges(frame, pixelsRadius)
	})

	t.Run("Empty pixels in radius", func(t *testing.T) {
		// Test with empty pixels array in PixelsRadius
		frame := createTestFrameForConstructor(3, 3, 100)
		pixelsRadius := []PixelsRadius{
			{
				OriginalX: 1,
				OriginalY: 1,
				CenterX:   0,
				CenterY:   0,
				Pixels:    Frame{}, // Empty pixels
			},
		}

		// This should panic when trying to access Pixels[CenterY][CenterX]
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with empty pixels array")
			}
		}()

		ApplyChanges(frame, pixelsRadius)
	})
}

func TestApplyChanges_Performance(t *testing.T) {
	t.Run("Large number of changes", func(t *testing.T) {
		// Create a larger frame for performance testing
		frame := createTestFrameForConstructor(100, 100, 128)
		
		// Create many pixel radius changes
		var pixelsRadius []PixelsRadius
		for y := 0; y < 100; y += 5 {
			for x := 0; x < 100; x += 5 {
				pixels := [][]uint8{{uint8((x + y) % 256)}}
				pixelsRadius = append(pixelsRadius, createTestPixelsRadius(x, y, 0, 0, pixels))
			}
		}

		// This should complete in reasonable time
		ApplyChanges(frame, pixelsRadius)

		// Verify some of the changes were applied
		if frame[0][0] != 0 {
			t.Errorf("Expected frame[0][0] to be 0, got %d", frame[0][0])
		}
		if frame[5][5] != 10 {
			t.Errorf("Expected frame[5][5] to be 10, got %d", frame[5][5])
		}
	})
}

// Benchmark for the ApplyChanges function
func BenchmarkApplyChanges(b *testing.B) {
	frame := createTestFrameForConstructor(50, 50, 100)
	pixelsRadius := []PixelsRadius{
		createTestPixelsRadius(10, 10, 1, 1, create3x3PixelRegion(150)),
		createTestPixelsRadius(20, 20, 1, 1, create3x3PixelRegion(200)),
		createTestPixelsRadius(30, 30, 1, 1, create3x3PixelRegion(250)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a fresh copy for each iteration
		frameCopy := make(Frame, len(frame))
		for j := range frame {
			frameCopy[j] = make([]uint8, len(frame[j]))
			copy(frameCopy[j], frame[j])
		}
		
		ApplyChanges(frameCopy, pixelsRadius)
	}
}
