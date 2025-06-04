package noise

import "math"

func Hash4(x, y, z, w uint32) (uint32, uint32, uint32, uint32) {
	// PCG4D
	x *= 1664525
	y *= 1664525
	z *= 1664525
	w *= 1664525
	x += 1013904223
	y += 1013904223
	z += 1013904223
	w += 1013904223
	x += y * w
	y += z * x
	z += x * y
	w += y * z
	x ^= x >> 16
	y ^= y >> 16
	z ^= z >> 16
	w ^= w >> 16
	x += y * w
	y += z * x
	z += x * y
	w += y * z
	return x, y, z, w
}

func Hash3(x, y, z uint32) (uint32, uint32, uint32) {
	// PCG3D
	x *= 1664525
	y *= 1664525
	z *= 1664525
	x += 1013904223
	y += 1013904223
	z += 1013904223
	x += y * z
	y += z * x
	z += x * y
	x ^= x >> 16
	y ^= y >> 16
	z ^= z >> 16
	x += y * z
	y += z * x
	z += x * y
	return x, y, z
}

func Hash2(x, y uint32) (uint32, uint32) {
	x, y, _ = Hash3(x, y, 0)
	return x, y
}

func Hash1(x uint32) uint32 {
	x, _ = Hash2(x, 0)
	return x
}

func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

func graddot2(seed, gridx, gridy uint32, fracx, fracy float64) float64 {
	h, _, _ := Hash3(seed, gridx, gridy)
	switch h & 7 {
	case 0:
		return fracx + fracy // (1, 1)
	case 1:
		return fracx - fracy // (1, -1)
	case 2:
		return -fracx + fracy // (-1, 1)
	case 3:
		return -fracx - fracy // (-1, -1)
	case 4:
		return fracx // (1, 0)
	case 5:
		return -fracx // (-1, 0)
	case 6:
		return fracy // (0, 1)
	case 7:
		return -fracy // (0, -1)
	}
	panic("unreachable")
}

// Sampler2 is a function that returns a value at position (x, y). The value range is [0, 1].
//
// The noise generation functions Perlin, Simplex, and the Fractal function for combining and modifying noise all return a Sampler,
// allowing them to be combined and used together.
type Sampler2 func(x, y float64) float64

func Perlin2(seed uint32) Sampler2 {
	return func(x, y float64) float64 {
		gridx := math.Floor(x)
		gridy := math.Floor(y)
		fracx := x - gridx
		fracy := y - gridy
		fadex := fade(fracx)
		fadey := fade(fracy)

		dot00 := graddot2(seed, uint32(gridx), uint32(gridy), fracx, fracy)
		dot01 := graddot2(seed, uint32(gridx+1), uint32(gridy), fracx-1, fracy)
		dot10 := graddot2(seed, uint32(gridx), uint32(gridy+1), fracx, fracy-1)
		dot11 := graddot2(seed, uint32(gridx+1), uint32(gridy+1), fracx-1, fracy-1)

		return lerp(
			lerp(dot00, dot01, fadex),
			lerp(dot10, dot11, fadex),
			fadey,
		)/2 + 0.5 // Normalize to [0, 1]
	}
}

func Fractal2(s Sampler2, octaves int, persistence, lacunarity float64, fns ...func(float64) float64) Sampler2 {
	return func(x, y float64) float64 {
		var total float64
		freq := 1.0
		amp := 1.0
		max := 0.0
		for range octaves {
			v := s(x*freq, y*freq)
			for _, fn := range fns {
				v = fn(v)
			}
			v *= amp
			total += v
			max += amp
			freq *= lacunarity
			amp *= persistence
		}
		return total / max
	}
}

func Turbulence(v float64) float64 {
	return math.Abs(v*2 - 1)
}

func Ridge(v float64) float64 {
	v = 1 - math.Abs(v*2-1)
	return v * v
}

func SmoothStep(edge0, edge1 float64) func(float64) float64 {
	if edge0 > edge1 {
		edge0, edge1 = edge1, edge0
	}
	return func(x float64) float64 {
		if x < edge0 {
			return 0
		} else if x > edge1 {
			return 1
		}
		t := (x - edge0) / (edge1 - edge0)
		return t * t * (3 - 2*t)
	}
}

// Bias adjusts the curve with a bias parameter.
// b = 0.5: no change, b < 0.5: bias towards 0, b > 0.5: bias towards 1
func Bias(b float64) func(float64) float64 {
	return func(v float64) float64 {
		return v / ((1/b-2)*(1-v) + 1)
	}
}

// Gain adjusts contrast symmetrically around 0.5.
// g = 0.5: no change, g < 0.5: increase contrast, g > 0.5: decrease contrast
func Gain(g float64) func(float64) float64 {
	return func(v float64) float64 {
		if v < 0.5 {
			return Bias(g)(v*2) / 2
		}
		return Bias(1-g)(v*2-1)/2 + 0.5
	}
}
