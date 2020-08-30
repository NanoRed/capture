package capture

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"math/rand"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

func init() {
	path := uio.GetBasePath()
	fontFile := []string{
		path + "Bodoni-16-Bold-11.ttf",
	}
	font = make([]*truetype.Font, 0)

	// Read the font data.
	for _, fileName := range fontFile {
		fontBytes, err := ioutil.ReadFile(fileName)
		if err != nil {
			appzaplog.Fatal("read fontFile error", zap.Error(err))
			panic(err)
		}
		tmp, err := freetype.ParseFont(fontBytes)
		if err != nil {
			appzaplog.Fatal("parse font error", zap.Error(err))
			panic(err)
		}
		font = append(font, tmp)
	}
}

type Captcha struct {
	Code  []byte
	Image []byte
}

var font []*truetype.Font

func (c *Captcha) genImage0() (err error) {

	var width int = 176
	var height int = 72
	var dpi float64 = 72
	var fontSize float64 = 50

	f_black := color.RGBA{0x2b, 0x2b, 0x2b, 0xff}
	b_white := color.RGBA{0xea, 0xea, 0xea, 0xff}

	fg := image.NewUniform(f_black)
	bg := image.NewUniform(b_white)
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), bg, image.Point{}, draw.Src)

	ctx := freetype.NewContext()
	ctx.SetDst(rgba)
	ctx.SetClip(rgba.Bounds())
	ctx.SetSrc(fg)
	ctx.SetDPI(dpi)
	ctx.SetFont(font[0])
	ctx.SetFontSize(fontSize)

	// draw the text
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

	// warp the image
	warp := image.NewRGBA(image.Rect(0, 0, rgba.Bounds().Dx(), rgba.Bounds().Dy()))
	draw.Draw(warp, warp.Bounds(), bg, image.Point{}, draw.Src)
	middle := width / 2
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(2)
	if n == 0 {
		n--
	}
	for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
		var y_ost int
		if x <= middle {
			y_ost = n * int(math.Log10(float64(middle-x+1))*5+0.5)
		} else {
			y_ost = -n * int(math.Log10(float64(x-middle+1))*5+0.5)
		}
		for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
			src_y := y + y_ost
			if src_y >= rgba.Bounds().Max.Y {
				src_y = rgba.Bounds().Max.Y - 1
			} else if src_y <= 0 {
				src_y = 1
			}
			warp.Set(x, y, rgba.At(x, src_y))
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
			if warp.At(x, y) == b_white {
				warp.Set(x, y, c)
			}
			if warp.At(x+1, y) == b_white {
				warp.Set(x+1, y, c)
			}
			if warp.At(x, y+1) == b_white {
				warp.Set(x, y+1, c)
			}
			if warp.At(x+1, y+1) == b_white {
				warp.Set(x+1, y+1, c)
			}
			x += x_advance + rand.Intn(3)
			y += y_advance + rand.Intn(3)
			if x <= 1 || x >= width-1 || y <= 1 || y >= height-1 {
				break
			}
		}
	}

	// turn it into byte
	buf := new(bytes.Buffer)
	err = png.Encode(buf, warp)
	if err != nil {
		return
	}
	c.Image = buf.Bytes()

	//imgBase64Str := base64.StdEncoding.EncodeToString(c.Image)
	//imgBase64Str = "data:image/png;base64," + imgBase64Str
	//log.Println(imgBase64Str)

	return
}

func (c *Captcha) genCode(n int) {
	var code bytes.Buffer
	b := []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'm', 'n', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'2', '3', '4', '5', '6', '7', '8', '9',
	}
	count := len(b)
	rand.Seed(time.Now().UnixNano())
	u := make(map[byte]struct{})
	for i := 0; i < n; i++ {
		r := b[rand.Intn(count-1)]
		_, ok := u[r]
		for ok {
			r = b[rand.Intn(count-1)]
			_, ok = u[r]
		}
		u[r] = struct{}{}
		code.WriteByte(r)
	}
	c.Code = code.Bytes()
}

func NewCaptcha(params ...interface{}) (c *Captcha, err error) {
	c = &Captcha{}
	defaultParams := []interface{}{4, 0} // [0]default 4 char [1]style 0
	for key, val := range params {
		defaultParams[key] = val
	}
	c.genCode(defaultParams[0].(int))
	switch defaultParams[1].(int) {
	case 0:
		err = c.genImage0()
	}
	return
}
