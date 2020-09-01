package wrapper

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"time"

	"github.com/RedAFD/SimpleCapture/attribute"
)

// DefaultWrapper default image wrapper, you can write your own wrapper to act on the image
func DefaultWrapper(attr *attribute.Attributes, input *image.RGBA) *image.RGBA {
	// drawing boards
	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// get background color
	bgColor := attr.BackGroundColor

	wrapper := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(wrapper, wrapper.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)

	// warping
	middle := width / 2
	rand.Seed(time.Now().UnixNano())
	sign := rand.Intn(2)
	if sign == 0 {
		sign--
	}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		var yOffset int
		correct := float64(height) / 72 * 5
		if x <= middle {
			yOffset = sign * int(math.Log10(float64(middle-x+1))*correct+0.5)
		} else {
			yOffset = -sign * int(math.Log10(float64(x-middle+1))*correct+0.5)
		}
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if srcY := y + yOffset; srcY >= bounds.Max.Y || srcY <= 0 {
				wrapper.Set(x, y, bgColor)
			} else if x == bounds.Min.X+1 && srcY == bounds.Max.Y-1 {
				wrapper.Set(x, y, bgColor)
			} else if x == bounds.Min.X+2 && srcY == bounds.Max.Y-1 {
				wrapper.Set(x, y, bgColor)
			} else {
				wrapper.Set(x, y, input.At(x, srcY))
			}
		}
	}

	// add some dot
	randColor := []color.RGBA{
		{0xd3, 0xb9, 0xb2, 0xff}, // red
		{0xba, 0xbe, 0xd4, 0xff}, // blue
		{0xc2, 0xb4, 0xc0, 0xff}, // purple
		{0xcf, 0xc0, 0xb8, 0xff}, // brown
		{0xbc, 0xd0, 0xce, 0xff}, // green
	}
	for i := 0; i < 35; i++ {
		// x && direction
		x := rand.Intn(width)
		xSign := rand.Intn(2)
		if xSign == 0 {
			xSign--
		}
		xAdvance := xSign * rand.Intn(4)
		// y && direction
		y := rand.Intn(height)
		ySign := rand.Intn(2)
		if ySign == 0 {
			ySign--
		}
		yAdvance := ySign * rand.Intn(4)
		// get rand color && draw
		c := randColor[rand.Intn(len(randColor))]
		for rand.Intn(15) > 0 {
			if wrapper.At(x, y) == bgColor {
				wrapper.Set(x, y, c)
			}
			if wrapper.At(x+1, y) == bgColor {
				wrapper.Set(x+1, y, c)
			}
			if wrapper.At(x, y+1) == bgColor {
				wrapper.Set(x, y+1, c)
			}
			if wrapper.At(x+1, y+1) == bgColor {
				wrapper.Set(x+1, y+1, c)
			}
			x += xAdvance + rand.Intn(3)
			y += yAdvance + rand.Intn(3)
			if x <= 1 || x >= width-1 || y <= 1 || y >= height-1 {
				break
			}
		}
	}

	return wrapper
}
