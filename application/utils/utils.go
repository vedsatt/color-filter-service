package utils

import (
	"image"
	"math"

	"git.miem.hse.ru/kg25-26/aisavelev.git/application/models"
)

func CalculateDominantColor(img image.Image) models.ColorRGBA {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	colorCount := make(map[models.ColorRGBA]int)
	totalPixels := 0

	// Попиксельная обработка
	step := 1
	if width > 1000 || height > 1000 {
		step = 2 // Для больших изображений
	}

	for y := 0; y < height; y += step {
		for x := 0; x < width; x += step {
			r, g, b, a := img.At(x, y).RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// Прозрачные пиксели не нужны
			if a8 > 0 {
				color := models.ColorRGBA{R: r8, G: g8, B: b8, A: a8}
				colorCount[color]++
				totalPixels++
			}
		}
	}

	// Самый частый цвет
	var dominantColor models.ColorRGBA
	maxCount := 0

	for color, count := range colorCount {
		if count > maxCount {
			maxCount = count
			dominantColor = color
		}
	}

	return dominantColor
}

func RGBToHSV(r, g, b uint8) [3]float64 {
	// Перевод в математический
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	// Разница между самым ярким и темным компонентом
	max := math.Max(math.Max(rf, gf), bf)
	min := math.Min(math.Min(rf, gf), bf)
	delta := max - min

	var h, s, v float64
	// Яркость V
	v = max

	// Насыщенность S
	if max != 0 {
		s = delta / max
	} else {
		return [3]float64{0, 0, 0}
	}

	// Оттенок H
	if delta == 0 {
		h = 0
	} else {
		switch max {
		case rf:
			h = (gf - bf) / delta
		case gf:
			h = 2 + (bf-rf)/delta
		case bf:
			h = 4 + (rf-gf)/delta
		}
		h *= 60
		if h < 0 {
			h += 360
		}
	}

	return [3]float64{h, s, v}
}

func ReverseSlice(slice []models.ImageAnalysis) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
