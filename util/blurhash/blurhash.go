package blurhash

// Vendor'd impl from https://github.com/bbrks/go-blurhash
/*
MIT License

Copyright (c) 2019 Ben Brooks

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
)

// ErrInvalidComponents is returned when components passed to Encode are invalid.
var ErrInvalidComponents = errors.New("blurhash: must have between 1 and 9 components")

// ErrInvalidHash is returned when the library encounters a hash it can't recognise.
var ErrInvalidHash = errors.New("blurhash: invalid hash")

func lengthError(expectedLength, actualLength int) error {
	// No stdlib support for wrapped errors, so return as-is pre-1.13
	return ErrInvalidHash
}

// Components returns the X and Y components of a blurhash.
func Components(hash string) (x, y int, err error) {
	sizeFlag, err := decode83(string(hash[0]))
	if err != nil {
		return 0, 0, err
	}

	x = (sizeFlag % 9) + 1
	y = (sizeFlag / 9) + 1

	expectedLength := 4 + 2*x*y
	actualLength := len(hash)
	if expectedLength != actualLength {
		return 0, 0, lengthError(expectedLength, actualLength)
	}

	return x, y, nil
}

// Decode returns an NRGBA image of the given hash with the given size.
func Decode(hash string, width, height int, punch int) (image.Image, error) {
	newImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	if err := DecodeDraw(newImg, hash, float64(punch)); err != nil {
		return nil, err
	}
	return newImg, nil
}

type drawImageNRGBA interface {
	SetNRGBA(x, y int, c color.NRGBA)
}

type drawImageRGBA interface {
	SetRGBA(x, y int, c color.RGBA)
}

// DecodeDraw decodes the given hash into the given image.
func DecodeDraw(dst draw.Image, hash string, punch float64) error {
	numX, numY, err := Components(hash)
	if err != nil {
		return err
	}

	quantisedMaximumValue, err := decode83(string(hash[1]))
	if err != nil {
		return err
	}
	maximumValue := float64(quantisedMaximumValue+1) / 166

	// for each component
	colors := make([][3]float64, numX*numY)
	for i := range colors {
		if i == 0 {
			val, err := decode83(hash[2:6])
			if err != nil {
				return err
			}
			colors[i] = decodeDC(val)
		} else {
			val, err := decode83(hash[4+i*2 : 6+i*2])
			if err != nil {
				return err
			}
			colors[i] = decodeAC(float64(val), maximumValue*punch)
		}
	}

	bounds := dst.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b float64

			for j := 0; j < numY; j++ {
				for i := 0; i < numX; i++ {
					basis := math.Cos(math.Pi*float64(x)*float64(i)/float64(width)) * math.Cos(math.Pi*float64(y)*float64(j)/float64(height))
					compColor := colors[i+j*numX]
					r += compColor[0] * basis
					g += compColor[1] * basis
					b += compColor[2] * basis
				}
			}

			sR := uint8(linearTosRGB(r))
			sG := uint8(linearTosRGB(g))
			sB := uint8(linearTosRGB(b))
			sA := uint8(255)

			// interface smuggle
			switch d := dst.(type) {
			case drawImageNRGBA:
				d.SetNRGBA(x, y, color.NRGBA{sR, sG, sB, sA})
			case drawImageRGBA:
				d.SetRGBA(x, y, color.RGBA{sR, sG, sB, sA})
			default:
				d.Set(x, y, color.NRGBA{sR, sG, sB, sA})
			}
		}
	}

	return nil
}

func decodeDC(val int) (c [3]float64) {
	c[0] = sRGBToLinear(val >> 16)
	c[1] = sRGBToLinear(val >> 8 & 255)
	c[2] = sRGBToLinear(val & 255)
	return c
}

func decodeAC(val, maximumValue float64) (c [3]float64) {
	quantR := math.Floor(val / (19 * 19))
	quantG := math.Mod(math.Floor(val/19), 19)
	quantB := math.Mod(val, 19)
	c[0] = signPow((quantR-9)/9, 2.0) * maximumValue
	c[1] = signPow((quantG-9)/9, 2.0) * maximumValue
	c[2] = signPow((quantB-9)/9, 2.0) * maximumValue
	return c
}

const (
	minComponents = 1
	maxComponents = 9
)

// Encode returns the blurhash for the given image.
func Encode(xComponents, yComponents int, img image.Image) (hash string, err error) {
	if xComponents < minComponents || xComponents > maxComponents ||
		yComponents < minComponents || yComponents > maxComponents {
		return "", ErrInvalidComponents
	}

	b := strings.Builder{}

	sizeFlag := (xComponents - 1) + (yComponents-1)*9
	sizeFlagEncoded, err := encode83(sizeFlag, 1)
	if err != nil {
		return "", err
	}

	_, err = b.WriteString(sizeFlagEncoded)
	if err != nil {
		return "", err
	}

	// vector of yComponents*xComponents*(RGB)
	factors := make([][][3]float64, yComponents)
	for y := 0; y < yComponents; y++ {
		factors[y] = make([][3]float64, xComponents)
		for x := 0; x < xComponents; x++ {
			factor := multiplyBasisFunction(x, y, img)
			factors[y][x][0] = factor[0]
			factors[y][x][1] = factor[1]
			factors[y][x][2] = factor[2]
		}
	}

	maximumValue := 0.0
	if xComponents*yComponents-1 > 0 {
		actualMaximumValue := 0.0
		for y := 0; y < yComponents; y++ {
			for x := 0; x < xComponents; x++ {
				if y == 0 && x == 0 {
					continue
				}
				actualMaximumValue = math.Max(math.Abs(factors[y][x][0]), actualMaximumValue)
				actualMaximumValue = math.Max(math.Abs(factors[y][x][1]), actualMaximumValue)
				actualMaximumValue = math.Max(math.Abs(factors[y][x][2]), actualMaximumValue)
			}
		}

		quantisedMaximumValue := math.Max(0, math.Min(82, math.Floor(actualMaximumValue*166-0.5)))
		maximumValue = (quantisedMaximumValue + 1) / 166
		str, err := encode83(int(quantisedMaximumValue), 1)
		if err != nil {
			return "", err
		}
		b.WriteString(str)
	} else {
		maximumValue = 1
		str, err := encode83(0, 1)
		if err != nil {
			return "", err
		}
		b.WriteString(str)
	}

	dc := factors[0][0]
	str, err := encode83(encodeDC(dc[0], dc[1], dc[2]), 4)
	if err != nil {
		return "", err
	}
	b.WriteString(str)

	for y := 0; y < yComponents; y++ {
		for x := 0; x < xComponents; x++ {
			if y == 0 && x == 0 {
				continue
			}
			str, err := encode83(encodeAC(factors[y][x][0], factors[y][x][1], factors[y][x][2], maximumValue), 2)
			if err != nil {
				return "", err
			}
			b.WriteString(str)
		}
	}

	return b.String(), nil
}

func encodeDC(r, g, b float64) int {
	return (linearTosRGB(r) << 16) + (linearTosRGB(g) << 8) + linearTosRGB(b)
}

func encodeAC(r, g, b, maximumValue float64) int {
	quantR := math.Max(0, math.Min(18, math.Floor(signPow(r/maximumValue, 0.5)*9+9.5)))
	quantG := math.Max(0, math.Min(18, math.Floor(signPow(g/maximumValue, 0.5)*9+9.5)))
	quantB := math.Max(0, math.Min(18, math.Floor(signPow(b/maximumValue, 0.5)*9+9.5)))

	return int(quantR*19*19 + quantG*19 + quantB)
}

func multiplyBasisFunction(xComponents, yComponents int, img image.Image) [3]float64 {
	var r, g, b float64
	width, height := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())

	normalisation := 2.0
	if xComponents == 0 && yComponents == 0 {
		normalisation = 1.0
	}

	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			//cR, cG, cB, _ := img.At(x, y).RGBA()
			c, ok := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if !ok {
				panic("not color.NRGBA")
			}
			basis := math.Cos(math.Pi*float64(xComponents)*float64(x)/width) *
				math.Cos(math.Pi*float64(yComponents)*float64(y)/height)
			r += basis * sRGBToLinear(int(c.R))
			g += basis * sRGBToLinear(int(c.G))
			b += basis * sRGBToLinear(int(c.B))
		}
	}

	scale := normalisation / (width * height)
	return [3]float64{
		r * scale,
		g * scale,
		b * scale,
	}
}

func signPow(val, exp float64) float64 {
	sign := 1.0
	if val < 0 {
		sign = -1
	}
	return sign * math.Pow(math.Abs(val), exp)
}

func sRGBToLinear(val int) float64 {
	v := float64(val) / 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func linearTosRGB(val float64) int {
	v := math.Max(0, math.Min(1, val))
	if v <= 0.0031308 {
		return int(v * 12.92 * 255 * 0.5)
	}
	return int((1.055*math.Pow(v, 1/2.4)-0.055)*255 + 0.5)
}

const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz#$%*+,-.:;=?@[]^_{|}~"

// Decode decodes a base83 string into an integer value.
func decode83(str string) (val int, err error) {
	for i, r := range str {
		idx := strings.IndexRune(chars, r)
		if idx == -1 {
			return 0, invalidError(r, i)
		}

		val = val*len(chars) + idx
	}
	return val, nil
}

// Encode encodes an integer value into a base83 string of the given length.
func encode83(val, length int) (str string, err error) {

	divisor := 1
	for i := 0; i < length-1; i++ {
		divisor *= len(chars)
	}

	for i := 0; i < length; i++ {
		idx := val / divisor % len(chars)
		divisor /= len(chars)
		str += string(chars[idx])
	}

	return str, nil
}

var ErrInvalidInput = errors.New("base83: invalid input")

func invalidError(r rune, i int) error {
	// No stdlib support for wrapped errors, so return as-is pre-1.13
	return ErrInvalidInput
}
