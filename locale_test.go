package validate

import (
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestI18n(t *testing.T) {
	va := Get(Printer(message.NewPrinter(language.Chinese)))
	err := va.Validate(map[string]int{
		"min": 2,
	}, Rules{
		".min": "min:5",
	})
	errm := err.Error()
	if err != nil && errm == "[invalid-data] `.min` 应该大于等于[5]，当前值为[2]" {
		return
	}
	t.Fatal("translate failed")
}
