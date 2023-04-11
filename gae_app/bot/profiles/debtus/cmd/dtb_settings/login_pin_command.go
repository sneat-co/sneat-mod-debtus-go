package dtb_settings

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"strconv"
	"strings"
)

var LoginPinCommand = botsfw.Command{
	Code: "LoginPin",
	Matcher: func(cmd botsfw.Command, whc botsfw.WebhookContext) bool {
		return false
		//if whc.BotPlatform().ID() == viber.PlatformID && whc.InputType() == botsfw.WebhookInputText {
		//	context := whc.Input().(viber.WebhookInputConversationStarted).GetContext()
		//	return strings.HasPrefix(context, "login-")
		//} else {
		//	return false
		//}
	},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		mt := whc.Input().(botsfw.WebhookTextMessage).Text()
		context := strings.Split(mt, " ")[0]
		contextParams := strings.Split(context, "_")
		var (
			loginID int64
			//gacID string
			lang string
		)
		if len(contextParams) < 2 || len(contextParams) > 3 {
			return m, fmt.Errorf("len(contextParams): %v", len(contextParams))
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
