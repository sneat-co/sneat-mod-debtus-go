package redirects

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/strongo/log"
	"google.golang.org/appengine"
)

func redirectToWebApp(w http.ResponseWriter, r *http.Request, authRequired bool, path string, p2p map[string]string, optionalParams []string) {
	c := appengine.NewContext(r)
	query := r.URL.Query()

	authInfo, _, err := auth.Authenticate(w, r, authRequired)
	if authRequired && err != nil {
		return
	}

	var redirectTo bytes.Buffer
	redirectTo.WriteString("/app/")

	lang := query.Get("lang")
	if lang == "" {
		if authInfo.UserID != 0 {
			user, err := dal.User.GetUserByID(c, authInfo.UserID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			lang = strings.ToLower(user.PreferredLocale()[:2])
		} else {
			lang = "en" // TODO: Bad to hard-code. Try to get from receipt?
		}
	}

	redirectTo.WriteString("#" + path)

	if path != "" {
		redirectTo.WriteString("&")
	}
	redirectTo.WriteString("lang=" + lang)

	sep := ""

	for pn, pn2 := range p2p {
		if pv := query.Get(pn); pv == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing required parameter: " + pn))
			return
		} else {
			pv = url.QueryEscape(pv)
			if pn == "id" && pn2 == "receipt" { // TODO: Dirty hack! Please fix!!!
				receiptID, err := common.DecodeID(pv)
				if err != nil {
					log.Debugf(c, "Failed to decode receipt ID: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("Failed to decod receipt ID: %v", err)))
					return
				}
				pv = strconv.FormatInt(receiptID, 10)
			}
			redirectTo.WriteString(sep + pn2 + "=" + pv)
		}
		sep = "&"
	}

	for _, p := range optionalParams {
		if v := query.Get(p); v != "" {
			redirectTo.WriteString(fmt.Sprintf("&%v=%v", p, url.QueryEscape(v)))
		}
	}

	if utm := query.Get("utm"); utm != "" {
		matches := reUtm.FindAllStringSubmatch(r.URL.RawQuery, -1) // TODO: Looks like a hack. Consider replacing ';' char with something else?
		if matches != nil && len(matches) == 1 {
			utm = matches[0][1]
			utmValues := strings.Split(utm, ";")
			if len(utmValues) == 3 {
				for i, p := range []string{"utm_source", "utm_medium", "utm_campaign"} {
					redirectTo.WriteString(fmt.Sprintf("&%v=%v", p, url.QueryEscape(utmValues[i])))
				}
			} else {
				log.Warningf(c, "Parameter utm should consist of 3 values seprated by ';' character. Got: [%v]", utm)
			}
		} else {
			log.Errorf(c, "reUtm: %v", matches)
		}
	} else {
		for _, p := range []string{"utm_source", "utm_medium", "utm_campaign"} {
			if v := query.Get(p); v != "" {
				redirectTo.WriteString(fmt.Sprintf("&%v=%v", p, url.QueryEscape(v)))
			}
		}
	}

	if authInfo.UserID > 0 {
		redirectTo.WriteString("&secret=" + query.Get("secret"))
	}
	log.Debugf(c, "Will redirect to: %v", redirectTo.String())
	http.Redirect(w, r, redirectTo.String(), http.StatusFound)
	//w.WriteHeader(http.StatusFound)
	//w.Header().Set("Location", redirectTo.String())
	//w.Write([]byte(fmt.Sprintf(`<html><head><meta http-equiv="refresh" content="0;URL='%v'" /></head></html>`, redirectTo.String())))
}

var reUtm = regexp.MustCompile(`(?:&|#|\?)?(?:utm=)(.+?)(?:&|#|$)`)
