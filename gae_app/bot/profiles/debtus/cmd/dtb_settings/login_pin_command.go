package dtb_settings

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/viber"
	"strconv"
	"strings"
)

var LoginPinCommand = bots.Command{
	Code: "LoginPin",
	Matcher: func(cmd bots.Command, whc bots.WebhookContext) bool {
		if whc.BotPlatform().Id() == viber_bot.ViberPlatformID && whc.InputType() == bots.WebhookInputText {
			context := whc.Input().(viber_bot.ViberWebhookInputConversationStarted).GetContext()
			return strings.HasPrefix(context, "login-")
		} else {
			return false
		}
	},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		mt := whc.Input().(bots.WebhookTextMessage).Text()
		context := strings.Split(mt, " ")[0]
		contextParams := strings.Split(context, "_")
		var (
			loginID int64
			//gacID string
			lang string
		)
		if len(contextParams) < 2 || len(contextParams) > 3 {
			return m, errors.New(fmt.Sprintf("len(contextParams): %v", len(contextParams)))
		}
		for _, p := range contextParams {
			switch {
			case strings.HasPrefix(p, "login-"):
				if loginID, err = strconv.ParseInt(p[len("login-"):], 10, 64); err != nil {
					err = errors.New(whc.Translate("Parameter 'login_id'  should be an integer."))
					return m, err
				}
			case strings.HasPrefix(p, "lang-"):
				lang = common.Locale2to5(p[len("lang-"):])
				whc.SetLocale(lang)
				whc.ChatEntity().SetPreferredLanguage(lang)
				//case strings.HasPrefix(p,"gac-"):
				//	gacID = p[len("gac-"):]
			}
		}
		c := whc.Context()
		if pinCode, err := facade.AuthFacade.AssignPinCode(c, loginID, whc.AppUserIntID()); err != nil {
			return m, err
		} else {
			return whc.NewMessage(fmt.Sprintf("Login PIN code: %v", pinCode)), nil
		}
	},
}
