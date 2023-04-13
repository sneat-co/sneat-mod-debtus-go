package gaedal

//import (
//	"context"
//	"github.com/strongo/app/gae"
//	"google.golang.org/appengine/v2/delay"
//)
//
//type TaskQueueDalGae struct {
//}
//
//func (TaskQueueDalGae) CallDelayFunc(c context.Context, queueName, subPath, key string, f interface{}, args ...interface{}) error {
//	return gae.CallDelayFunc(c, queueName, subPath, delay.MustRegister(key, f), args...)
//}
