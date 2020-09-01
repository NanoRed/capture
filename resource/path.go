package resource

/*
char* resourcePath()
{
    return __FILE__;
}
*/
import "C"
import (
	"bytes"
	"path"
)

var resourcePath string

func init() {
	resourcePath = path.Dir(C.GoString(C.resourcePath()))
}

// GetResourcePath get resource path
func GetResourcePath() string {
	return resourcePath
}

// GetResourceFontFile get font file path
func GetResourceFontFile(filename string) string {
	var buffer bytes.Buffer
	buffer.WriteString(resourcePath)
	buffer.WriteString("/font/")
	buffer.WriteString(filename)

	return buffer.String()
}
