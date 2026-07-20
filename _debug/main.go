package main

import (
	"fmt"
	"math"
)

const NakshatraSpan = 360.0 / 27.0

func main() {
	tests := []float64{0, 13.33, 13.34, 26.67, 40, 80, 120, 160, 200, 240, 280, 320, 6.67, 10}
	for _, lon := range tests {
		sidLon := math.Mod(lon, 360)
		idx := int((sidLon + 1e-9) / NakshatraSpan) % 27
		degInNak := sidLon - float64(idx)*NakshatraSpan
		if degInNak < 0 {
			degInNak += NakshatraSpan
		}
		if degInNak > NakshatraSpan-1e-9 {
			degInNak = 0
			idx = (idx + 1) % 27
		}
		pada := int(degInNak/(NakshatraSpan/4.0)) + 1
		if pada > 4 {
			pada = 4
		}
		fmt.Printf("lon=%.2f idx=%d degInNak=%.6f pada=%d\n", lon, idx, degInNak, pada)
	}

	fmt.Println()
	navTests := []float64{0, 3.34, 10, 29.9, 30, 33.34, 40, 60, 63.34, 90, 93.34}
	for _, lon := range navTests {
		sidLon := math.Mod(lon, 360)
		signIdx := int(sidLon / 30.0) % 12
		degInSign := sidLon - float64(signIdx)*30.0
		if degInSign < 0 {
			degInSign += 30.0
		}
		if degInSign > 30.0-1e-9 {
			degInSign = 0
		}
		padaNum := int(degInSign / (30.0 / 9.0))
		if padaNum > 8 {
			padaNum = 8
		}
		fmt.Printf("lon=%.2f signIdx=%d degInSign=%.6f padaNum=%d\n", lon, signIdx, degInSign, padaNum)
	}
}
