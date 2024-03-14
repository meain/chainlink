package main

type HSL struct {
	H, S, L float64
}

type RGB struct {
	R, G, B float64
}

func HSLToRGB(hsl HSL) RGB {
	var r, g, b float64

	if hsl.S == 0 {
		r = hsl.L
		g = hsl.L
		b = hsl.L
	} else {
		var q float64
		if hsl.L < 0.5 {
			q = hsl.L * (1 + hsl.S)
		} else {
			q = hsl.L + hsl.S - (hsl.L * hsl.S)
		}
		p := 2*hsl.L - q
		hk := hsl.H / 360.0

		// Convert hue to RGB
		r = hueToRGB(p, q, hk+1/3.0)
		g = hueToRGB(p, q, hk)
		b = hueToRGB(p, q, hk-1/3.0)
	}

	return RGB{r, g, b}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1/2.0 {
		return q
	}
	if t < 2/3.0 {
		return p + (q-p)*6*(2/3.0-t)
	}
	return p
}
