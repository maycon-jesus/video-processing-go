package internal

import (
	"math"
	"testing"
)

// Helper function to create a test frame
func createTestFrame(height, width int, value uint8) Frame {
	frame := make(Frame, height)
	for i := range frame {
		frame[i] = make([]uint8, width)
		for j := range frame[i] {
			frame[i][j] = value
		}
	}
	return frame
}

// Helper function to create a frame with specific pattern
func createPatternFrame(height, width int) Frame {
	frame := make(Frame, height)
	for i := range frame {
		frame[i] = make([]uint8, width)
		for j := range frame[i] {
			frame[i][j] = uint8((i + j) % 256)
		}
	}
	return frame
}

// Helper function to create a frame with edge pattern
func createEdgeFrame() Frame {
	frame := make(Frame, 5)
	for i := range frame {
		frame[i] = make([]uint8, 5)
	}

	// Create a simple edge pattern
	frame[0] = []uint8{0, 0, 0, 0, 0}
	frame[1] = []uint8{0, 0, 0, 0, 0}
	frame[2] = []uint8{255, 255, 255, 255, 255}
	frame[3] = []uint8{255, 255, 255, 255, 255}
	frame[4] = []uint8{255, 255, 255, 255, 255}

	return frame
}

// Helper function to create a noisy frame
func createNoisyFrame() Frame {
	frame := make(Frame, 5)
	for i := range frame {
		frame[i] = make([]uint8, 5)
	}

	// Create a frame with one noisy pixel
	frame[0] = []uint8{100, 100, 100, 100, 100}
	frame[1] = []uint8{100, 100, 100, 100, 100}
	frame[2] = []uint8{100, 100, 255, 100, 100} // Noisy pixel at center
	frame[3] = []uint8{100, 100, 100, 100, 100}
	frame[4] = []uint8{100, 100, 100, 100, 100}

	return frame
}

func TestGetPixelRadius(t *testing.T) {
	tests := []struct {
		name     string
		frame    Frame
		y, x     int
		radius   int
		expected PixelsRadius
	}{
		{
			name:   "Normal case - center of frame",
			frame:  createTestFrame(5, 5, 100),
			y:      2,
			x:      2,
			radius: 1,
			expected: PixelsRadius{
				CenterX:   1,
				CenterY:   1,
				OriginalX: 2,
				OriginalY: 2,
				XMin:      1,
				XMax:      3,
				YMin:      1,
				YMax:      3,
			},
		},
		{
			name:   "Edge case - top-left corner",
			frame:  createTestFrame(5, 5, 50),
			y:      0,
			x:      0,
			radius: 1,
			expected: PixelsRadius{
				CenterX:   0,
				CenterY:   0,
				OriginalX: 0,
				OriginalY: 0,
				XMin:      0,
				XMax:      1,
				YMin:      0,
				YMax:      1,
			},
		},
		{
			name:   "Edge case - bottom-right corner",
			frame:  createTestFrame(5, 5, 200),
			y:      4,
			x:      4,
			radius: 1,
			expected: PixelsRadius{
				CenterX:   1,
				CenterY:   1,
				OriginalX: 4,
				OriginalY: 4,
				XMin:      3,
				XMax:      4,
				YMin:      3,
				YMax:      4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPixelRadius(tt.frame, tt.y, tt.x, tt.radius)

			if result.CenterX != tt.expected.CenterX {
				t.Errorf("CenterX = %d, expected %d", result.CenterX, tt.expected.CenterX)
			}
			if result.CenterY != tt.expected.CenterY {
				t.Errorf("CenterY = %d, expected %d", result.CenterY, tt.expected.CenterY)
			}
			if result.OriginalX != tt.expected.OriginalX {
				t.Errorf("OriginalX = %d, expected %d", result.OriginalX, tt.expected.OriginalX)
			}
			if result.OriginalY != tt.expected.OriginalY {
				t.Errorf("OriginalY = %d, expected %d", result.OriginalY, tt.expected.OriginalY)
			}
			if result.XMin != tt.expected.XMin {
				t.Errorf("XMin = %d, expected %d", result.XMin, tt.expected.XMin)
			}
			if result.XMax != tt.expected.XMax {
				t.Errorf("XMax = %d, expected %d", result.XMax, tt.expected.XMax)
			}
			if result.YMin != tt.expected.YMin {
				t.Errorf("YMin = %d, expected %d", result.YMin, tt.expected.YMin)
			}
			if result.YMax != tt.expected.YMax {
				t.Errorf("YMax = %d, expected %d", result.YMax, tt.expected.YMax)
			}

			// Check that pixels are correctly copied
			if len(result.Pixels) == 0 {
				t.Error("Pixels should not be empty")
			}
		})
	}
}

func TestPixelsRadius_IsEdgePixel(t *testing.T) {
	tests := []struct {
		name      string
		pixels    PixelsRadius
		threshold float64
		expected  bool
	}{
		{
			name: "Edge pixel detected",
			pixels: PixelsRadius{
				CenterX: 1,
				CenterY: 1,
				Pixels:  createEdgeFrame(),
			},
			threshold: 25.0,
			expected:  true,
		},
		{
			name: "Uniform region - no edge",
			pixels: PixelsRadius{
				CenterX: 2,
				CenterY: 2,
				Pixels:  createTestFrame(5, 5, 100),
			},
			threshold: 25.0,
			expected:  false,
		},
		{
			name: "Small region - considered edge",
			pixels: PixelsRadius{
				CenterX: 0,
				CenterY: 0,
				Pixels:  createTestFrame(2, 2, 100),
			},
			threshold: 25.0,
			expected:  true,
		},
		{
			name: "Center at boundary - considered edge",
			pixels: PixelsRadius{
				CenterX: 0,
				CenterY: 2,
				Pixels:  createTestFrame(5, 5, 100),
			},
			threshold: 25.0,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pixels.IsEdgePixel(tt.threshold)
			if result != tt.expected {
				t.Errorf("IsEdgePixel() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestPixelsRadius_CalculateVariance(t *testing.T) {
	tests := []struct {
		name     string
		pixels   PixelsRadius
		expected float64
		delta    float64
	}{
		{
			name: "Uniform pixels - zero variance",
			pixels: PixelsRadius{
				Pixels: createTestFrame(3, 3, 100),
			},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name: "Empty pixels",
			pixels: PixelsRadius{
				Pixels: Frame{},
			},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name: "High variance pattern",
			pixels: PixelsRadius{
				Pixels: Frame{
					{0, 255, 0},
					{255, 0, 255},
					{0, 255, 0},
				},
			},
			expected: 16055.555556, // Calculated expected variance: mean=127.5, variance=16055.555556
			delta:    0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pixels.CalculateVariance()
			if math.Abs(result-tt.expected) > tt.delta {
				t.Errorf("CalculateVariance() = %f, expected %f (±%f)", result, tt.expected, tt.delta)
			}
		})
	}
}

func TestPixelsRadius_IsNoisePixel(t *testing.T) {
	tests := []struct {
		name     string
		pixels   PixelsRadius
		expected bool
	}{
		{
			name: "Noisy pixel detected",
			pixels: PixelsRadius{
				CenterX: 2,
				CenterY: 2,
				Pixels:  createNoisyFrame(),
			},
			expected: true,
		},
		{
			name: "Uniform region - not noise",
			pixels: PixelsRadius{
				CenterX: 2,
				CenterY: 2,
				Pixels:  createTestFrame(5, 5, 100),
			},
			expected: false,
		},
		{
			name: "Similar neighbors - not noise",
			pixels: PixelsRadius{
				CenterX: 1,
				CenterY: 1,
				Pixels: Frame{
					{100, 105, 98},
					{102, 100, 103},
					{99, 101, 97},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pixels.IsNoisePixel()
			if result != tt.expected {
				t.Errorf("IsNoisePixel() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestPixelsRadius_applyMedianFilter(t *testing.T) {
	pixels := PixelsRadius{
		CenterX: 1,
		CenterY: 1,
		Pixels: Frame{
			{100, 105, 98},
			{102, 255, 103}, // Center pixel is 255 (noise)
			{99, 101, 97},
		},
	}

	neighbors := []uint8{100, 105, 98, 102, 103, 99, 101, 97}
	result := pixels.applyMedianFilter(neighbors)

	// The median of the neighbors should be around 100-101
	if result < 98 || result > 103 {
		t.Errorf("applyMedianFilter() = %d, expected value between 98-103", result)
	}
}

func TestPixelsRadius_applyMeanFilter(t *testing.T) {
	pixels := PixelsRadius{
		CenterX: 1,
		CenterY: 1,
		Pixels: Frame{
			{100, 100, 100},
			{100, 200, 100}, // Center pixel is 200
			{100, 100, 100},
		},
	}

	neighbors := []uint8{100, 100, 100, 100, 100, 100, 100, 100}
	current := uint8(200)
	alpha := 0.5

	result := pixels.applyMeanFilter(neighbors, current, alpha)
	expected := uint8(0.5*100 + 0.5*200) // Should be 150

	if result != expected {
		t.Errorf("applyMeanFilter() = %d, expected %d", result, expected)
	}
}

func TestPixelsRadius_applySoftFilter(t *testing.T) {
	pixels := PixelsRadius{
		CenterX: 1,
		CenterY: 1,
		Pixels: Frame{
			{100, 100, 100},
			{100, 200, 100}, // Center pixel is 200
			{100, 100, 100},
		},
	}

	neighbors := []uint8{100, 100, 100, 100, 100, 100, 100, 100}
	current := uint8(200)
	alpha := 0.3

	result := pixels.applySoftFilter(neighbors, current, alpha)
	expected := uint8(0.3*100 + 0.7*200) // Should be 170

	if result != expected {
		t.Errorf("applySoftFilter() = %d, expected %d", result, expected)
	}
}

func TestPixelsRadius_ApplyAdaptiveFilter(t *testing.T) {
	tests := []struct {
		name     string
		pixels   PixelsRadius
		testFunc func(t *testing.T, original, filtered PixelsRadius)
	}{
		{
			name: "Apply filter to edge pixel",
			pixels: PixelsRadius{
				CenterX: 2,
				CenterY: 2,
				Pixels:  createEdgeFrame(),
			},
			testFunc: func(t *testing.T, original, filtered PixelsRadius) {
				// Edge pixels should be preserved or only slightly modified
				originalValue := original.Pixels[original.CenterY][original.CenterX]
				filteredValue := filtered.Pixels[filtered.CenterY][filtered.CenterX]
				diff := math.Abs(float64(originalValue) - float64(filteredValue))
				if diff > 50 { // Allow some change but not too much
					t.Errorf("Edge pixel changed too much: %d -> %d", originalValue, filteredValue)
				}
			},
		},
		{
			name: "Apply filter to noisy pixel",
			pixels: PixelsRadius{
				CenterX: 2,
				CenterY: 2,
				Pixels:  createNoisyFrame(),
			},
			testFunc: func(t *testing.T, original, filtered PixelsRadius) {
				// Noisy pixels should be significantly changed
				originalValue := original.Pixels[original.CenterY][original.CenterX]
				filteredValue := filtered.Pixels[filtered.CenterY][filtered.CenterX]
				if originalValue == 255 && filteredValue == 255 {
					t.Error("Noisy pixel should have been filtered")
				}
			},
		},
		{
			name: "Apply filter to uniform region",
			pixels: PixelsRadius{
				CenterX: 2,
				CenterY: 2,
				Pixels:  createTestFrame(5, 5, 100),
			},
			testFunc: func(t *testing.T, original, filtered PixelsRadius) {
				// Uniform regions should remain mostly unchanged
				originalValue := original.Pixels[original.CenterY][original.CenterX]
				filteredValue := filtered.Pixels[filtered.CenterY][filtered.CenterX]
				if originalValue != filteredValue {
					// Small change is acceptable due to filtering
					diff := math.Abs(float64(originalValue) - float64(filteredValue))
					if diff > 10 {
						t.Errorf("Uniform pixel changed too much: %d -> %d", originalValue, filteredValue)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy to compare before and after
			original := PixelsRadius{
				CenterX:   tt.pixels.CenterX,
				CenterY:   tt.pixels.CenterY,
				OriginalX: tt.pixels.OriginalX,
				OriginalY: tt.pixels.OriginalY,
				XMin:      tt.pixels.XMin,
				XMax:      tt.pixels.XMax,
				YMin:      tt.pixels.YMin,
				YMax:      tt.pixels.YMax,
				Pixels:    make(Frame, len(tt.pixels.Pixels)),
			}

			for i := range tt.pixels.Pixels {
				original.Pixels[i] = make([]uint8, len(tt.pixels.Pixels[i]))
				copy(original.Pixels[i], tt.pixels.Pixels[i])
			}

			// Apply the filter
			tt.pixels.ApplyAdaptiveFilter()

			// Run the test function
			tt.testFunc(t, original, tt.pixels)
		})
	}
}

func TestPixelsRadius_ApplyAdaptiveFilter_EmptyPixels(t *testing.T) {
	pixels := PixelsRadius{
		Pixels: Frame{},
	}

	// Should not panic with empty pixels
	pixels.ApplyAdaptiveFilter()
}

func TestFilterBoundaryConditions(t *testing.T) {
	// Test with minimum size frame
	frame := Frame{{255}}
	radius := GetPixelRadius(frame, 0, 0, 1)

	// Should not panic
	radius.ApplyAdaptiveFilter()

	// Test various edge cases
	testCases := []struct {
		name  string
		frame Frame
		y, x  int
		r     int
	}{
		{"Single pixel", Frame{{100}}, 0, 0, 0},
		{"2x2 frame center", Frame{{100, 150}, {200, 250}}, 0, 0, 1},
		{"Large radius", createTestFrame(3, 3, 100), 1, 1, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test %s panicked: %v", tc.name, r)
				}
			}()

			radius := GetPixelRadius(tc.frame, tc.y, tc.x, tc.r)
			radius.ApplyAdaptiveFilter()
		})
	}
}

// Benchmark tests
func BenchmarkGetPixelRadius(b *testing.B) {
	frame := createTestFrame(100, 100, 128)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetPixelRadius(frame, 50, 50, 5)
	}
}

func BenchmarkIsEdgePixel(b *testing.B) {
	frame := createEdgeFrame()
	pixels := PixelsRadius{
		CenterX: 2,
		CenterY: 2,
		Pixels:  frame,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pixels.IsEdgePixel(25.0)
	}
}

func BenchmarkCalculateVariance(b *testing.B) {
	pixels := PixelsRadius{
		Pixels: createPatternFrame(10, 10),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pixels.CalculateVariance()
	}
}

func BenchmarkApplyAdaptiveFilter(b *testing.B) {
	frame := createPatternFrame(10, 10)
	pixels := PixelsRadius{
		CenterX: 5,
		CenterY: 5,
		Pixels:  frame,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a fresh copy for each iteration since the filter modifies the data
		testPixels := PixelsRadius{
			CenterX: pixels.CenterX,
			CenterY: pixels.CenterY,
			Pixels:  make(Frame, len(pixels.Pixels)),
		}
		for j := range pixels.Pixels {
			testPixels.Pixels[j] = make([]uint8, len(pixels.Pixels[j]))
			copy(testPixels.Pixels[j], pixels.Pixels[j])
		}
		testPixels.ApplyAdaptiveFilter()
	}
}
