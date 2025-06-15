package internal

import (
	"math"
	"reflect"
	"testing"
)

func TestMedian(t *testing.T) {
	tests := []struct {
		name     string
		values   []uint8
		expected uint8
	}{
		{
			name:     "odd number of elements",
			values:   []uint8{1, 3, 5, 7, 9},
			expected: 5,
		},
		{
			name:     "even number of elements",
			values:   []uint8{2, 4, 6, 8},
			expected: 6,
		},
		{
			name:     "single element",
			values:   []uint8{42},
			expected: 42,
		},
		{
			name:     "unsorted values",
			values:   []uint8{9, 1, 5, 3, 7},
			expected: 5,
		},
		{
			name:     "duplicate values",
			values:   []uint8{1, 1, 2, 2, 3},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy to ensure original is not modified
			original := make([]uint8, len(tt.values))
			copy(original, tt.values)

			result := median(tt.values)

			if result != tt.expected {
				t.Errorf("median(%v) = %d, expected %d", tt.values, result, tt.expected)
			}

			// Verify original slice is not modified
			if !reflect.DeepEqual(tt.values, original) {
				t.Errorf("original slice was modified: got %v, expected %v", tt.values, original)
			}
		})
	}
}

func TestIsEdgePixel(t *testing.T) {
	// Create test video frames
	videoFrames := make(VideoFrames, 3)
	for i := range videoFrames {
		videoFrames[i] = createTestFrame(5, 5, 100)
	}

	// Create a strong edge pattern in the middle frame with higher contrast
	// Set up a clear horizontal edge
	for j := 0; j < 5; j++ {
		videoFrames[1][0][j] = 0   // Top row dark
		videoFrames[1][1][j] = 50  // Second row medium
		videoFrames[1][2][j] = 100 // Middle row medium
		videoFrames[1][3][j] = 150 // Fourth row medium-bright
		videoFrames[1][4][j] = 255 // Bottom row bright
	}

	tests := []struct {
		name         string
		currentFrame int
		line         int
		pixel        int
		expected     bool
	}{
		{
			name:         "border pixel - top edge",
			currentFrame: 1,
			line:         0,
			pixel:        2,
			expected:     true,
		},
		{
			name:         "border pixel - left edge",
			currentFrame: 1,
			line:         2,
			pixel:        0,
			expected:     true,
		},
		{
			name:         "border pixel - right edge",
			currentFrame: 1,
			line:         2,
			pixel:        4,
			expected:     true,
		},
		{
			name:         "border pixel - bottom edge",
			currentFrame: 1,
			line:         4,
			pixel:        2,
			expected:     true,
		},
		{
			name:         "center pixel with strong gradient",
			currentFrame: 1,
			line:         2,
			pixel:        2,
			expected:     true, // Should detect edge due to strong vertical gradient
		},
		{
			name:         "pixel with weak gradient",
			currentFrame: 0, // Use frame with uniform values
			line:         2,
			pixel:        2,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEdgePixel(videoFrames, tt.currentFrame, tt.line, tt.pixel)
			if result != tt.expected {
				t.Errorf("isEdgePixel(frame=%d, line=%d, pixel=%d) = %v, expected %v",
					tt.currentFrame, tt.line, tt.pixel, result, tt.expected)
			}
		})
	}
}

func TestIsBlur(t *testing.T) {
	tests := []struct {
		name     string
		values   []uint8
		current  uint8
		expected bool
	}{
		{
			name:     "insufficient values",
			values:   []uint8{100},
			current:  50,
			expected: false,
		},
		{
			name:     "blur detected - low current value with high median",
			values:   []uint8{150, 140, 160, 145},
			current:  50,
			expected: true,
		},
		{
			name:     "no blur - similar values",
			values:   []uint8{100, 105, 95, 110},
			current:  102,
			expected: false,
		},
		{
			name:     "no blur - high current value",
			values:   []uint8{150, 140, 160, 145},
			current:  200,
			expected: false,
		},
		{
			name:     "no blur - small difference",
			values:   []uint8{100, 105, 95, 110},
			current:  80,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBlur(tt.values, tt.current)
			if result != tt.expected {
				t.Errorf("isBlur(%v, %d) = %v, expected %v",
					tt.values, tt.current, result, tt.expected)
			}
		})
	}
}

func TestIsFlare(t *testing.T) {
	tests := []struct {
		name     string
		values   []uint8
		current  uint8
		expected bool
	}{
		{
			name:     "insufficient values",
			values:   []uint8{100},
			current:  200,
			expected: false,
		},
		{
			name:     "flare detected - high current value with low median",
			values:   []uint8{50, 60, 40, 55},
			current:  200,
			expected: true,
		},
		{
			name:     "no flare - similar values",
			values:   []uint8{100, 105, 95, 110},
			current:  102,
			expected: false,
		},
		{
			name:     "no flare - low current value",
			values:   []uint8{50, 60, 40, 55},
			current:  70,
			expected: false,
		},
		{
			name:     "no flare - small difference",
			values:   []uint8{150, 160, 140, 155},
			current:  200,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFlare(tt.values, tt.current)
			if result != tt.expected {
				t.Errorf("isFlare(%v, %d) = %v, expected %v",
					tt.values, tt.current, result, tt.expected)
			}
		})
	}
}

func TestIsNoise(t *testing.T) {
	tests := []struct {
		name     string
		values   []uint8
		current  uint8
		variance float64
		expected bool
	}{
		{
			name:     "insufficient values",
			values:   []uint8{100},
			current:  120,
			variance: 10.0,
			expected: false,
		},
		{
			name:     "noise detected - stable previous values with outlier current",
			values:   []uint8{100, 102, 98, 101, 99},
			current:  130,
			variance: 5.0,
			expected: true,
		},
		{
			name:     "no noise - unstable previous values",
			values:   []uint8{50, 150, 75, 125, 80},
			current:  130,
			variance: 30.0,
			expected: false,
		},
		{
			name:     "no noise - current value close to median",
			values:   []uint8{100, 102, 98, 101, 99},
			current:  103,
			variance: 5.0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNoise(tt.values, tt.current, tt.variance)
			if result != tt.expected {
				t.Errorf("isNoise(%v, %d, %f) = %v, expected %v",
					tt.values, tt.current, tt.variance, result, tt.expected)
			}
		})
	}
}

func TestHasMovement(t *testing.T) {
	tests := []struct {
		name     string
		values   []uint8
		expected bool
	}{
		{
			name:     "insufficient values",
			values:   []uint8{100},
			expected: false,
		},
		{
			name:     "high variance - movement detected",
			values:   []uint8{50, 150, 75, 200, 25},
			expected: true,
		},
		{
			name:     "low variance - no movement",
			values:   []uint8{100, 102, 98, 101, 99},
			expected: false,
		},
		{
			name:     "medium variance - borderline case",
			values:   []uint8{80, 120, 90, 110, 85},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasMovement(tt.values)
			if result != tt.expected {
				t.Errorf("hasMovement(%v) = %v, expected %v",
					tt.values, result, tt.expected)
			}
		})
	}
}

func TestAdaptiveTemporalFilter(t *testing.T) {
	tests := []struct {
		name     string
		values   []uint8
		current  uint8
		variance float64
		expected uint8
	}{
		{
			name:     "no previous values",
			values:   []uint8{},
			current:  100,
			variance: 0,
			expected: 100,
		},
		{
			name:     "low variance - high weight to median",
			values:   []uint8{100, 102, 98, 101, 99},
			current:  150,
			variance: 5.0,
			expected: 120, // 0.6 * 100 + 0.4 * 150 = 120
		},
		{
			name:     "medium variance - medium weight",
			values:   []uint8{80, 120, 90, 110, 85},
			current:  150,
			variance: 15.0,
			expected: 126, // 0.4 * 90 + 0.6 * 150 = 126
		},
		{
			name:     "high variance - low weight to median",
			values:   []uint8{50, 200, 75, 175, 100},
			current:  150,
			variance: 40.0,
			expected: 140, // 0.2 * 100 + 0.8 * 150 = 140
		},
		{
			name:     "result clipping - lower bound",
			values:   []uint8{10, 15, 12, 8, 11},
			current:  0,
			variance: 5.0,
			expected: 7, // 0.6 * 11 + 0.4 * 0 = 6.6 ≈ 7
		},
		{
			name:     "result clipping - upper bound",
			values:   []uint8{240, 245, 250, 248, 242},
			current:  255,
			variance: 5.0,
			expected: 249, // 0.6 * 245 + 0.4 * 255 = 147 + 102 = 249
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adaptiveTemporalFilter(tt.values, tt.current, tt.variance)
			// Allow small rounding differences
			if abs(int(result)-int(tt.expected)) > 1 {
				t.Errorf("adaptiveTemporalFilter(%v, %d, %f) = %d, expected %d",
					tt.values, tt.current, tt.variance, result, tt.expected)
			}
		})
	}
}

func TestCalculateVariance(t *testing.T) {
	tests := []struct {
		name     string
		values   []uint8
		expected float64
	}{
		{
			name:     "empty slice",
			values:   []uint8{},
			expected: 0,
		},
		{
			name:     "single value",
			values:   []uint8{100},
			expected: 0,
		},
		{
			name:     "identical values",
			values:   []uint8{100, 100, 100, 100},
			expected: 0,
		},
		{
			name:     "simple case",
			values:   []uint8{1, 2, 3, 4, 5},
			expected: 2,
		},
		{
			name:     "high variance",
			values:   []uint8{0, 255, 0, 255},
			expected: 16256.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateVariance(tt.values)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("calculateVariance(%v) = %f, expected %f",
					tt.values, result, tt.expected)
			}
		})
	}
}

func TestTimeTravalerProcessLine(t *testing.T) {
	// Create test video frames
	videoFrames := make(VideoFrames, 10)
	for i := range videoFrames {
		videoFrames[i] = createTestFrame(5, 5, uint8(100+i))
	}

	tests := []struct {
		name           string
		currentFrame   int
		previousFrames int
		line           int
		expectedLength int
	}{
		{
			name:           "early frame - no processing",
			currentFrame:   1,
			previousFrames: 3,
			line:           2,
			expectedLength: 5,
		},
		{
			name:           "processable frame",
			currentFrame:   5,
			previousFrames: 3,
			line:           2,
			expectedLength: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeTravalerProcessLine(videoFrames, tt.currentFrame, tt.previousFrames, tt.line)

			if len(result) != tt.expectedLength {
				t.Errorf("TimeTravalerProcessLine() returned line with length %d, expected %d",
					len(result), tt.expectedLength)
			}

			// For early frames, result should be identical to original
			if tt.currentFrame <= 2 {
				originalLine := videoFrames[tt.currentFrame][tt.line]
				if !reflect.DeepEqual(result, originalLine) {
					t.Errorf("Early frame processing should return original line")
				}
			}
		})
	}
}

func TestTimeTravaler(t *testing.T) {
	// Create test video frames
	videoFrames := make(VideoFrames, 10)
	for i := range videoFrames {
		videoFrames[i] = createTestFrame(3, 3, uint8(100+i))
	}

	tests := []struct {
		name           string
		currentFrame   int
		previousFrames int
		shouldProcess  bool
	}{
		{
			name:           "insufficient previous frames",
			currentFrame:   2,
			previousFrames: 3,
			shouldProcess:  false,
		},
		{
			name:           "sufficient previous frames",
			currentFrame:   5,
			previousFrames: 3,
			shouldProcess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy to compare
			originalFrame := make(Frame, len(videoFrames[tt.currentFrame]))
			for i := range originalFrame {
				originalFrame[i] = make([]uint8, len(videoFrames[tt.currentFrame][i]))
				copy(originalFrame[i], videoFrames[tt.currentFrame][i])
			}

			TimeTravaler(videoFrames, tt.currentFrame, tt.previousFrames)

			// Check if frame was modified as expected
			frameModified := !reflect.DeepEqual(videoFrames[tt.currentFrame], originalFrame)

			if tt.shouldProcess && !frameModified {
				t.Error("Frame should have been processed but wasn't modified")
			} else if !tt.shouldProcess && frameModified {
				t.Error("Frame should not have been processed but was modified")
			}
		})
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests
func BenchmarkMedian(b *testing.B) {
	values := []uint8{9, 1, 5, 3, 7, 2, 8, 4, 6}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		median(values)
	}
}

func BenchmarkCalculateVarianceTimeTraveler(b *testing.B) {
	values := []uint8{100, 105, 95, 110, 98, 103, 97, 108, 102, 99}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateVariance(values)
	}
}

func BenchmarkTimeTravalerProcessLine(b *testing.B) {
	// Create test video frames
	videoFrames := make(VideoFrames, 10)
	for i := range videoFrames {
		videoFrames[i] = createTestFrame(100, 100, uint8(100+i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TimeTravalerProcessLine(videoFrames, 5, 3, 50)
	}
}
