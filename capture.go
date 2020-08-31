package capture

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type Captcha struct {
	Code  []byte
	Image []byte
}

func (c *Captcha) String() string {
	imgBase64Str := base64.StdEncoding.EncodeToString(c.Image)
	return fmt.Sprintf("\nCode:\n\t%s\nBase64Image:\n\tdata:image/png;base64,%s", c.Code, imgBase64Str)
}

func (c *Captcha) generateCode(charCount int) {
	var code bytes.Buffer
	dict := []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'm', 'n', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'2', '3', '4', '5', '6', '7', '8', '9',
	}
	count := len(dict)
	rand.Seed(time.Now().UnixNano())
	code.WriteByte('m')
	for i := 0; i < charCount-1; i++ {
		code.WriteByte(dict[rand.Intn(count-1)])
	}
	c.Code = code.Bytes()
}

func (c *Captcha) generateImage(width, height int, fontSize float64, fontHandler *truetype.Font, wrapper func(input *image.RGBA) *image.RGBA) (err error) {

	black := color.RGBA{0x2b, 0x2b, 0x2b, 0xff}
	whiteSmoke := color.RGBA{0xea, 0xea, 0xea, 0xff}

	pencil := image.NewUniform(black)
	background := image.NewUniform(whiteSmoke)
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), background, image.Point{}, draw.Src)

	ctx := freetype.NewContext()
	ctx.SetDst(rgba)
	ctx.SetClip(rgba.Bounds())
	ctx.SetSrc(pencil)
	ctx.SetDPI(72)
	ctx.SetFont(fontHandler)
	ctx.SetFontSize(fontSize)

	// draw the code text
	size := int(ctx.PointToFixed(fontSize) >> 6)
	xSpace := size - 20
	x_offset := (width - len(c.Code)*xSpace) / 2
	y_offset := size
	for _, v := range c.Code {
		_, err = ctx.DrawString(string(v), freetype.Pt(x_offset, y_offset))
		if err != nil {
			return
		}
		x_offset += xSpace
	}

	// record background and pencil color
	rgba.Set(rgba.Bounds().Min.X+1, rgba.Bounds().Max.Y-1, whiteSmoke)
	rgba.Set(rgba.Bounds().Min.X+2, rgba.Bounds().Max.Y-1, black)

	if wrapper == nil {
		wrapper = defaultWrapper
	}
	rgba = wrapper(rgba)

	// turn it into byte
	buf := new(bytes.Buffer)
	err = png.Encode(buf, rgba)
	if err != nil {
		return
	}
	c.Image = buf.Bytes()

	return
}

func defaultWrapper(input *image.RGBA) *image.RGBA {
	// drawing board attribute
	attr := input.Bounds()
	width := attr.Dx()
	height := attr.Dy()

	// get background and pencil color
	bgColor := input.At(attr.Min.X+1, attr.Max.Y-1)
	//pcColor := input.At(attr.Min.X+1, attr.Max.Y)

	wrapper := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(wrapper, wrapper.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)

	// warping
	middle := width / 2
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(2)
	if n == 0 {
		n--
	}
	for x := attr.Min.X; x < attr.Max.X; x++ {
		var y_ost int
		correct := float64(height)/72*5
		if x <= middle {
			y_ost = n * int(math.Log10(float64(middle-x+1))*correct+0.5)
		} else {
			y_ost = -n * int(math.Log10(float64(x-middle+1))*correct+0.5)
		}
		for y := attr.Min.Y; y < attr.Max.Y; y++ {
			if src_y := y + y_ost; src_y >= attr.Max.Y || src_y <= 0 {
				wrapper.Set(x, y, bgColor)
			} else if x == attr.Min.X+1 && src_y == attr.Max.Y-1 {
				wrapper.Set(x, y, bgColor)
			} else if x == attr.Min.X+2 && src_y == attr.Max.Y-1 {
				wrapper.Set(x, y, bgColor)
			} else {
				wrapper.Set(x, y, input.At(x, src_y))
			}
		}
	}

	// add some dot
	rand_color := []color.RGBA{
		{0xd3, 0xb9, 0xb2, 0xff}, // red
		{0xba, 0xbe, 0xd4, 0xff}, // blue
		{0xc2, 0xb4, 0xc0, 0xff}, // purple
		{0xcf, 0xc0, 0xb8, 0xff}, // brown
		{0xbc, 0xd0, 0xce, 0xff}, // green
	}
	for i := 0; i < 35; i++ {
		// x && direction
		x := rand.Intn(width)
		x_pre := rand.Intn(2)
		if x_pre == 0 {
			x_pre--
		}
		x_advance := x_pre * rand.Intn(4)
		// y && direction
		y := rand.Intn(height)
		y_pre := rand.Intn(2)
		if y_pre == 0 {
			y_pre--
		}
		y_advance := y_pre * rand.Intn(4)
		// get rand color && draw
		c := rand_color[rand.Intn(len(rand_color))]
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
			x += x_advance + rand.Intn(3)
			y += y_advance + rand.Intn(3)
			if x <= 1 || x >= width-1 || y <= 1 || y >= height-1 {
				break
			}
		}
	}

	return wrapper
}

var fontHandler = &struct {
	sync.Map
	storeLock sync.Mutex
}{}

type CaptureAttributes struct {
	Width int
	Height int
	FontFile string
	FontSize float64
	CharCount int
	Wrapper func(input *image.RGBA) *image.RGBA

	fontHandler *truetype.Font
}

func New(customAttr... *CaptureAttributes) (capture *Captcha, err error) {
	// default attributes
	attr := &CaptureAttributes{
		Width:     176,
		Height:    72,
		FontFile:  "./font/Bodoni-16-Bold-11.ttf",
		FontSize:  50,
		CharCount: 4,
		Wrapper:   nil,
	}
	if len(customAttr) > 0 {
		attr = customAttr[0]
	}

	// font handler cache
	if handler, ok := fontHandler.Load(attr.FontFile); ok {
		attr.fontHandler = handler.(*truetype.Font)
	} else {
		fontHandler.storeLock.Lock()
		defer fontHandler.storeLock.Unlock()
		// double check
		if handler, ok := fontHandler.Load(attr.FontFile); ok {
			attr.fontHandler = handler.(*truetype.Font)
		} else {
			var fontFileBytes []byte
			fontFileBytes, err = ioutil.ReadFile(attr.FontFile)
			if err != nil {
				return
			}
			attr.fontHandler, err = freetype.ParseFont(fontFileBytes)
			if err != nil {
				return
			}
			fontHandler.Store(attr.FontFile, attr.fontHandler)
		}
	}

	// generate capture
	capture = new(Captcha)
	capture.generateCode(attr.CharCount)
	err = capture.generateImage(attr.Width, attr.Height, attr.FontSize, attr.fontHandler, attr.Wrapper)

	return
}
