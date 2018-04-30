package website

import (
	"fmt"
	"net/http"
	"strconv"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/website/admin"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/website/pages"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/website/redirects"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	//"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	//"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/api"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/website/pages/inspector"
	"github.com/julienschmidt/httprouter"
)

func InitWebsite(router *httprouter.Router) {
	router.GET("/", pages.IndexRootPage)

	redirects.InitRedirects(router)

	for _, locale := range strongo.LocalesByCode5 {
		localeSiteCode := locale.SiteCode()
		router.GET(fmt.Sprintf("/%v/ads", localeSiteCode), pages.AdsPage)
		router.GET(fmt.Sprintf("/%v/help-us", localeSiteCode), pages.HelpUsPage)
		router.GET(fmt.Sprintf("/%v/login", localeSiteCode), LoginHandler)
		router.GET(fmt.Sprintf("/%v/counterparty", localeSiteCode), pages.CounterpartyPage)
		router.GET(fmt.Sprintf("/%v/", localeSiteCode), pages.IndexPage)
		//strongo.AddHTTPHandler(fmt.Sprintf("/%v/create-mass-invite", localeSiteCode), api.AuthOnly(CreateInvitePage))

	}
	router.GET("/en/songs/annie-iou-a-dance", pages.AnnieIOUaDancePage)
	router.GET("/en/songs/iou-by-dappy", pages.IOWDappyPage)

	admin.InitAdmin(router)
	inspector.InitInspector(router)
}

func CreateInvitePage(w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	c := appengine.NewContext(r)
	log.Infof(c, "CreateInvitePage()")
	//panic("Not implemented")
	userID := authInfo.UserID
	//session, _ := common.GetSession(r)
	//userID := session.GetUserID()
	//if userID == 0 {
	//	w.WriteHeader(http.StatusUnauthorized)
	//	return
	//}
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "templates/create-mass-invite.html")
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		inviteCode := r.Form.Get("Code")
		if !dal.InviteCodeRegex.Match([]byte(inviteCode)) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Invate code [%v] does not match pattern: %v", inviteCode, dal.InviteCodeRegex.String())))
			return
		}
		maxClaimsCount, err := strconv.ParseInt(r.Form.Get("MaxClaimsCount"), 10, 32)
		if err != nil || inviteCode == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		invite, err := dal.Invite.GetInvite(c, inviteCode)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if invite != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Invate code [%v] already exists", inviteCode)))
			return
		}
		translator := strongo.NewSingleMapTranslator(strongo.GetLocaleByCode5(strongo.LocaleCodeEnUS), strongo.NewMapTranslator(c, trans.TRANS))
		ec := strongo.NewExecutionContext(c, translator)
		if _, err = dal.Invite.CreateMassInvite(ec, userID, inviteCode, int32(maxClaimsCount), "web"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(fmt.Sprintf("Invite created, code: %v, MaxClaimsCount: %v", inviteCode, maxClaimsCount)))
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
