package models

import (
	"image/color"
	"time"
)

type ColorInfo struct {
	R, G, B     uint8
	H, S, V     float64
	FileName    string
	Deviation   float64
	HueDistance float64
}

type ImageAnalysis struct {
	FileName    string
	DominantRGB ColorRGBA
	HSV         [3]float64
	Deviation   float64
	IsTarget    bool
}

type Config struct {
	Tolerance     float64
	SortDirection string
	TargetColor   ColorRGBA
}

type Session struct {
	ID        string
	Images    []ImageAnalysis
	CreatedAt time.Time
}

type ColorRGBA color.RGBA
