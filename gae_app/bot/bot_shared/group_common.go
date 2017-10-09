package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bytes"
	"strconv"
)

func GroupCallbackCommandData(command string, groupID string) string {
	var b bytes.Buffer
	b.WriteString(command)
	b.WriteString("?group=")
	return string(strconv.AppendInt(b.Bytes(), groupID, 10))
}

func GroupCallbackCommand(code string, f func(whc bots.WebhookContext, callbackURL *url.URL, group models.Group) (m bots.MessageFromBot, err error)) bots.Command {
	return bots.NewCallbackCommand(code,
		func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
			var group models.Group
			if group, err = GetGroup(whc); err != nil {
				return
			}
			return f(whc, callbackURL, group)
		},
	)
}
