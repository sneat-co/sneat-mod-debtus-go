package reminders

import "github.com/strongo/delaying"

func InitDelaying(mustRegisterFunc func(key string, i any) delaying.Delayer) {
	delaySetChatIsForbidden = mustRegisterFunc("SetChatIsForbidden", SetChatIsForbidden)
}

var (
	delaySetChatIsForbidden delaying.Delayer
)
