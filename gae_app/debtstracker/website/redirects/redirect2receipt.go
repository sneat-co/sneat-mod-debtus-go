package redirects

import (
	"fmt"
	"github.com/strongo/i18n"
	"google.golang.org/appengine/v2"
	"html/template"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/common"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/website/pages"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
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
	receiptID := receiptCode
	if receiptID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var err error
	log.Debugf(c, "Receipt ID: %v", receiptID)
	_, err = dtdal.Receipt.GetReceiptByID(c, nil, receiptID)
	switch err {
	case nil: //pass
	case datastore.ErrNoSuchEntity:
		log.Debugf(c, "Receipt not found by ID")
		http.NotFound(w, r)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(c, err.Error())
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	//lang := query.Get("lang")
	//if lang == "" {
	//	lang = receipt.Data.Lang
	//}

	if strings.HasPrefix(r.UserAgent(), "facebookexternalhit/") || query.Get("for") == "fb" {
		if receiptOpenGraphPageTmpl == nil {
			receiptOpenGraphPageTmpl = template.Must(template.ParseFiles(pages.TEMPLATES_PATH + "receipt-opengraph.html"))
		}
		locale := i18n.LocaleEnUS // strongo.GetLocaleByCode5(receipt.Lang) // TODO: Check for empty
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
