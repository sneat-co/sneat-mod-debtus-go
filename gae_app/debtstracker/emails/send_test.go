package emails

import (
	"testing"
	"context"
)

func TestGetEmailTextWithoutTranslator(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should fail")
		}
	}()
	c := context.Background()
	GetEmailText(c, nil, "some-template", nil)
}

func TestGetEmailHtmlWithoutTranslator(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should fail")
		}
	}()
	c := context.Background()
	GetEmailHtml(c, nil, "some-template", nil)
}
