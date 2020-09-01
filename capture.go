package capture

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/RedAFD/SimpleCapture/attribute"
	"github.com/RedAFD/SimpleCapture/resource"
	"github.com/RedAFD/SimpleCapture/wrapper"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"time"
)

type Captcha struct {
	Code  []byte
	Image []byte

	attributes *attribute.Attributes
}

func (c *Captcha) String() string {
	imgBase64Str := base64.StdEncoding.EncodeToString(c.Image)
	return fmt.Sprintf("\nCode:\n\t%s\nBase64Image:\n\tdata:image/png;base64,%s", c.Code, imgBase64Str)
}

func (c *Captcha) Reload() (err error) {
	c.generateCode()
	err = c.generateImage()
	return
}

func (c *Captcha) generateCode() {
	var code bytes.Buffer
	dict := []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'm', 'n', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'2', '3', '4', '5', '6', '7', '8', '9',
	}
	count := len(dict)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < c.attributes.CharCount; i++ {
		code.WriteByte(dict[rand.Intn(count-1)])
	}
	c.Code = code.Bytes()
}

func (c *Captcha) generateImage() (err error) {

	textUF := image.NewUniform(c.attributes.CharColor)
	backgroundUF := image.NewUniform(c.attributes.BackGroundColor)
	rgba := image.NewRGBA(image.Rect(0, 0, c.attributes.Width, c.attributes.Height))
	draw.Draw(rgba, rgba.Bounds(), backgroundUF, image.Point{}, draw.Src)

	ftc := freetype.NewContext()
	ftc.SetDst(rgba)
	ftc.SetClip(rgba.Bounds())
	ftc.SetSrc(textUF)
	ftc.SetDPI(72)
	ftc.SetFont(c.attributes.FontHandler)
	ftc.SetFontSize(c.attributes.FontSize)

	// calculate the widths and print to image
	face := truetype.NewFace(c.attributes.FontHandler, &truetype.Options{Size: c.attributes.FontSize})
	charWidth := make(map[byte]int)
	textWidth := 0
	for _, x := range c.Code {
		rawAdvance, ok := face.GlyphAdvance(rune(x))
		if !ok {
			err = errors.New("can't get advance")
			return
		}
		advance := int(float64(rawAdvance) / 64)
		charWidth[x] = advance
		textWidth += advance
	}
	xOffset := (c.attributes.Width - textWidth) / 2
	yOffset := int(ftc.PointToFixed(c.attributes.FontSize) >> 6)
	for _, x := range c.Code {
		_, err = ftc.DrawString(string(x), freetype.Pt(xOffset, yOffset))
		if err != nil {
			return
		}
		xOffset += charWidth[x]
	}

	if c.attributes.Wrapper != nil {
		rgba = c.attributes.Wrapper(c.attributes, rgba)
	}

	// turn it into byte
	buf := new(bytes.Buffer)
	err = png.Encode(buf, rgba)
	if err != nil {
		return
	}
	c.Image = buf.Bytes()

	return
}

func New(attributes... *attribute.Attributes) (capture *Captcha, err error) {

	// init capture attribute
	var attr *attribute.Attributes
	if last := len(attributes)-1; last >= 0 {
		attr = attributes[last]
	} else {
		// default attribute
		attr = &attribute.Attributes{
			Width:           176,
			Height:          72,
			FontFile:        resource.ResourceFontFile("Bodoni-16-Bold-11.ttf"),
			FontSize:        50,
			CharCount:       4,
			CharColor:       color.RGBA{0x2b, 0x2b, 0x2b, 0xff},
			BackGroundColor: color.RGBA{0xea, 0xea, 0xea, 0xff},
			Wrapper:         wrapper.DefaultWrapper,
		}
	}
	if attr.FontHandler == nil {
		err = attr.CreateFontHandler()
		if err != nil {
			return
		}
	}

	// generate capture
	capture = new(Captcha)
	capture.attributes = attr
	err = capture.Reload()

	return
}
