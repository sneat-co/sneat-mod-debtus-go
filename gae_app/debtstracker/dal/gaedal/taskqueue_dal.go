package gaedal

//import (
//	"golang.org/x/net/context"
//	"github.com/strongo/app/gae"
//	"google.golang.org/appengine/delay"
//)
//
//type TaskQueueDalGae struct {
//}
//
//func (_ TaskQueueDalGae) CallDelayFunc(c context.Context, queueName, subPath, key string, f interface{}, args ...interface{}) error {
//	return gae.CallDelayFunc(c, queueName, subPath, delay.Func(key, f), args...)
//}
