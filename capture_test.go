package capture

import "testing"

func TestCapture(t *testing.T) {
	capture, err := New()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(capture)
		t.Log(capture.attributes.FontFile)
	}
	err = capture.Reload()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(capture)
		t.Log(capture.attributes.FontFile)
	}
}