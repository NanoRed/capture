package attribute

import (
	"sync"
	"io/ioutil"
	"image"
	"image/color"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type Attributes struct {
	Width int
	Height int
	FontFile string
	FontSize float64
	FontHandler *truetype.Font
	CharCount int
	CharColor color.RGBA
	BackGroundColor color.RGBA
	Wrapper func(attr *Attributes, input *image.RGBA) *image.RGBA
}

// font handler cache
var fontHandler = &struct {
	sync.Map
	storeLock sync.Mutex
}{}

func (a *Attributes) CreateFontHandler() (err error) {
	if handler, ok := fontHandler.Load(a.FontFile); ok {
		a.FontHandler = handler.(*truetype.Font)
	} else {
		fontHandler.storeLock.Lock()
		defer fontHandler.storeLock.Unlock()
		// double check
		if handler, ok := fontHandler.Load(a.FontFile); ok {
			a.FontHandler = handler.(*truetype.Font)
		} else {
			var fontFileBytes []byte
			fontFileBytes, err = ioutil.ReadFile(a.FontFile)
			if err != nil {
				return
			}
			a.FontHandler, err = freetype.ParseFont(fontFileBytes)
			if err != nil {
				return
			}
			fontHandler.Store(a.FontFile, a.FontHandler)
		}
	}
	return
}