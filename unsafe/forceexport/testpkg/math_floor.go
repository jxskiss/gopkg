package testpkg

import "math"

func Floor(x float64) float64 {
	return floor(x)
}

func floor(x float64) float64 {
	if x == 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}
	if x < 0 {
		d, fract := math.Modf(-x)
		if fract != 0.0 {
			d = d + 1
		}
		return -d
	}
	d, _ := math.Modf(x)
	return d
}
