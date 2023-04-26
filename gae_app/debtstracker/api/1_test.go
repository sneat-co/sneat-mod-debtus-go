package api

import (
	"github.com/strongo/app/delaying"
)

func init() {
	delaying.Init(delaying.VoidWithLog)
	InitDelaying(delaying.MustRegisterFunc)
}
