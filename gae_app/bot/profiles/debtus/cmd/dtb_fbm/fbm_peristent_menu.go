package dtb_fbm

import (
	"net/http"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-fbm"
	"strings"
	"github.com/DebtsTracker/translations/emoji"
	"fmt"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
)

var EM_SPACE = strings.Repeat("\u00A0", 2)


func SetPersistentMenu(c context.Context, r *http.Request, bot bots.BotSettings, api fbm_api.GraphAPI) (err error) {
	url := fmt.Sprintf("https://%v/app/#fbm%v", r.Host, bot.ID)

	//menuItemWebUrl := func(icon, title, hash string) fbm_api.MenuItemWebUrl {
	//	return fbm_api.NewMenuItemWebUrl(
	//		icon + EM_SPACE + title,
	//		url + hash, fbm_api.WebviewHeightRatioFull, false, true)
	//}
	menuItemPostback := func(icon, title, payload string) fbm_api.MenuItemPostback {
		return fbm_api.NewMenuItemPostback(icon + EM_SPACE+ title, payload)
	}

	log.Debugf(c, "url: %v", url)

	persistentMenu := func(locale string) fbm_api.PersistentMenu {

		//topMenuDebts := fbm_api.NewMenuItemNested(emoji.MEMO_ICON + EM_SPACE + "Debts",
		//	menuItemWebUrl(emoji.TAKE_ICON, "I borrowed", "#new-debt=contact-to-user"),
		//	menuItemWebUrl(emoji.GIVE_ICON, "I lent", "#new-debt=user-to-contact"),
		//	menuItemWebUrl(emoji.RETURN_BACK_ICON, "Returned", "#debt-returned"),
		//	menuItemWebUrl(emoji.BALANCE_ICON, "Balance", "#debts"),
		//)
		//
		//topMenuBills := fbm_api.NewMenuItemNested(emoji.BILLS_ICON + " Bills",
		//	menuItemWebUrl(emoji.DIVIDE_ICON, "Split bill", "#split-bill"),
		//	menuItemWebUrl(emoji.MONEY_BAG_ICON, "Start collecting", "#start-collecting"),
		//	menuItemWebUrl(emoji.OPEN_BOOK_ICON, "Outstanding bills", "#bills=outstanding"),
		//	menuItemWebUrl(emoji.CALENDAR_ICON, "Recurring bills", "#bills=recurring"),
		//)
		//
		//topMenuView := fbm_api.NewMenuItemNested(emoji.TOTAL_ICON + EM_SPACE + "More...",
		//	menuItemPostback(emoji.HOME_ICON, "Get started", "fbm-get-started"),
		//	menuItemWebUrl(emoji.CONTACTS_ICON, "Contacts", "#contacts"),
		//	menuItemWebUrl(emoji.HISTORY_ICON, "History", "#history"),
		//	menuItemWebUrl(emoji.SETTINGS_ICON, "Settings", "#settings"),
		//)
		//
		//return fbm_api.NewPersistentMenu(locale, false,
		//	topMenuDebts,
		//	topMenuBills,
		//	topMenuView,
		//)
		return fbm_api.NewPersistentMenu(locale, false,
			menuItemPostback(emoji.HOME_ICON, "Main menu", FbmMainMenuCommand.Code),
			menuItemPostback(emoji.MEMO_ICON, "Debt", FbmDebtsCommand.Code),
			menuItemPostback(emoji.BILLS_ICON, "Bills", FbmBillsCommand.Code),
		)
	}

	persistentMenuMessage := fbm_api.PersistentMenuMessage{
		PersistentMenus: []fbm_api.PersistentMenu{
			persistentMenu("default"),
			//persistentMenu("ru_RU"),
		},
	}

	if err = api.SetPersistentMenu(c, persistentMenuMessage); err != nil {
		return
	}
	return
}
