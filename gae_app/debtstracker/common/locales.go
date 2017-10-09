package common

import (
	"fmt"
	"github.com/strongo/app"
	"strings"
)

func Locale2to5(locale2 string) string {
	if len(locale2) != 2 {
		panic("len(locale2) != 2")
	}
	if strings.ToLower(locale2) == "en" {
		return strongo.LOCALE_EN_US
	} else {
		return fmt.Sprintf("%v-%v", strings.ToLower(locale2), strings.ToUpper(locale2))
	}
}
