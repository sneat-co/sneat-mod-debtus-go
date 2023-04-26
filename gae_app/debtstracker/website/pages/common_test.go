package pages

import (
	"github.com/strongo/app"
	"testing"
)

func TestRenderCachedPageWithoutArguemnts(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic")
		}
	}()
	RenderCachedPage(nil, nil, nil, strongo.LocaleEnUS, nil, 0)
}
