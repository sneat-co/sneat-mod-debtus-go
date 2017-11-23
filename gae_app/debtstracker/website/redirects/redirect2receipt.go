package redirects

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/website/pages"
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"html/template"
	"net/http"
	"strings"
	"github.com/julienschmidt/httprouter"
)

var receiptOpenGraphPageTmpl *template.Template

func ReceiptRedirect(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)
	query := r.URL.Query()
	receiptCode := query.Get("id")
	if receiptCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	receiptID, err := common.DecodeID(receiptCode)
	if err != nil || receiptID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Debugf(c, "Receipt ID: %v", receiptID)
	receipt, err := dal.Receipt.GetReceiptByID(c, receiptID)
	switch err {
	case nil: //pass
	case datastore.ErrNoSuchEntity:
		log.Debugf(c, "Receipt not found by ID")
		http.NotFound(w, r)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(c, err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	lang := query.Get("lang")
	if lang == "" {
		lang = receipt.Lang
	}

	if strings.HasPrefix(r.UserAgent(), "facebookexternalhit/") || query.Get("for") == "fb" {
		if receiptOpenGraphPageTmpl == nil {
			receiptOpenGraphPageTmpl = template.Must(template.ParseFiles(pages.TEMPLATES_PATH + "receipt-opengraph.html"))
		}
		locale := strongo.LocaleEnUS // strongo.GetLocaleByCode5(receipt.Lang) // TODO: Check for empty
		pages.RenderCachedPage(w, r, receiptOpenGraphPageTmpl, locale, map[string]interface{}{
			"host":      r.Host,
			"ogUrl":     r.URL.String(),
			"ReceiptID": receiptID,
			//"ReceiptCode": common.EncodeID(receiptID),
			"Title":       fmt.Sprintf("Receipt @ DebtsTracker.io #%v", receiptID),
			"Description": "Receipt description goes here",
		}, 9)
	} else {
		redirectToWebApp(w, r, false, common.Deeplink.AppHashPathToReceipt(receiptID), map[string]string{}, []string{})
	}
}
