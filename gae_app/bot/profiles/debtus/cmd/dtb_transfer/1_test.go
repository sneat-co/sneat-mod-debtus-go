package dtb_transfer

import (
	"github.com/strongo/app/delaying"
)

func init() {
	delaying.Init(delaying.VoidWithLog)
	InitDelaying(delaying.MustRegisterFunc)
}
