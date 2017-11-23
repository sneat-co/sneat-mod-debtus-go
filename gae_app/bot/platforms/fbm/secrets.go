package fbm

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/log"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/fbm"
	fb "github.com/strongo/facebook"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

const (
	fbm_PROD_PAGE_ACCESS_TOKEN  = "CAAGdsU6rzXgBAIVVFCQKsmZCue0PUpzuZA4BaZA80UfhxnRH2Nbf5Ri9K66tkXwkLuPa2WhN53MsiAngUUcE2wZBuhb5ZBO0DV5hVAQbOFuCuL5rP35FFQuf2NCkSs0IVwmhpkXkeAAt3a4yn4ZCnBkfearPByE4gbvSD4WfswvZBb6GrtTJ2ZAEgvawDfUWKKdcm8yXsuz2ZBAZDZD"
	fbm_TEST_PAGE_ACCESS_TOKEN  = "EAAIOtyFmtbsBAA4CuLiZALf4R4voPZBg3AySB63XB8SdRsid7FB2dWwHJgAgONJ0olWEGcOEVYXEjsZBeQ1M124keNAWhgWj3XwIDJ4mfYCl1m1DUuwaZCaOZCm7BZCY6TWwAKTRL5Uv0BSilWVhwGZBDcVmUg8Cm5na19KrFUVOAZDZD"
	fbm_LOCAL_PAGE_ACCESS_TOKEN = "EAAGgOpi8kU8BADaJYHciKTZAgVnvewxXZCGVsoeRZBHN2o3mXj1sEXsQwfrQMGHZCGprtRY61TRwpX15ZAqYRXmbMpDjPhcts4M1fnWwWbZA9ZCVYEAW0htPj6XFZAX1AWHOlm9CirI8qx2G5k9Hg62hK7VZCNdvZCtPICAj8BwbAnegZDZD"

	fbm_SPLITBILL_PROD_PAGE_ACCESS_TOKEN = "EAAFzhuSvVPIBAIo4R2oBtrhfHEXOYUOqA4I6ogBqkHhOeJmb6SDWtptIcAe40J0Nbliqwh32omjjPQrxbPydsnwGZBxCEx7QEfQXEGsfs9JLLKFlZCqEDeO35pAoZCLriDRVjIAc6oMbxWOwMgNd4xxdJfSio88okNxo88imAZDZD"
)

type fbAppSecrets struct {
	AppID     string
	AppSecret string
	app       *fb.App
}

func (s *fbAppSecrets) App() *fb.App {
	if s.app == nil {
		s.app = fb.New(s.AppID, s.AppSecret)
	}
	return s.app
}

var (
	fbLocal = fbAppSecrets{
		AppID:     "457648507752783",
		AppSecret: "23ceb7a7f53516119fd60b19a309cb14",
	}
	fbDev = fbAppSecrets{
		AppID:     "579129655604667",
		AppSecret: "0e3ee2d65e8abae458f121e874950b73",
	}
	fbProd = fbAppSecrets{
		AppID:     "454859831364984",
		AppSecret: "72f6f7382dda3235d48e6a7d60bb4a6a",
	}
)

var _bots bots.SettingsBy

func Bots(_ context.Context) bots.SettingsBy {
	if len(_bots.ByCode) == 0 {
		_bots = bots.NewBotSettingsBy(nil,
			fbm_bot.NewFbmBot(
				strongo.EnvProduction,
				bot.ProfileDebtus,
				"debtstracker",
				"1587055508253137",
				fbm_PROD_PAGE_ACCESS_TOKEN,
				"d6087a01-c728-4fdf-983c-1695d76236dc",
				trans.SupportedLocalesByCode5[strongo.LOCALE_EN_US],
			),
			fbm_bot.NewFbmBot(
				strongo.EnvProduction,
				bot.ProfileSplitus,
				"splitbill.co",
				"286238251784027",
				fbm_SPLITBILL_PROD_PAGE_ACCESS_TOKEN,
				"e8535dd1-df3b-4c3f-bd2c-d4a822509bb3",
				trans.SupportedLocalesByCode5[strongo.LOCALE_EN_US],
			),
			fbm_bot.NewFbmBot(
				strongo.EnvDevTest,
				bot.ProfileDebtus,
				"debtstracker.dev",
				"942911595837341",
				fbm_TEST_PAGE_ACCESS_TOKEN,
				"4afb645e-b592-48e6-882c-89f0ec126fbb",
				trans.SupportedLocalesByCode5[strongo.LOCALE_EN_US],
			),
			fbm_bot.NewFbmBot(
				strongo.EnvLocal,
				bot.ProfileDebtus,
				"debtstracker.local",
				"300392587037950",
				fbm_LOCAL_PAGE_ACCESS_TOKEN,
				"4afb645e-b592-48e6-882c-89f0ec126fbb",
				trans.SupportedLocalesByCode5[strongo.LOCALE_EN_US],
			),
		)
	}
	return _bots
}

var ErrUnknownHost = errors.New("Unknown host")

func GetFbAppAndHost(r *http.Request) (fbApp *fb.App, host string, err error) {
	switch r.Host {
	case "debtstracker.io":
		return fbProd.App(), r.Host, nil
	case "debtstracker-io.appspot.com":
		return fbProd.App(), "debtstracker.io", nil
	case "debtstracker-dev1.appspot.com":
		return fbDev.App(), r.Host, nil
	case "debtstracker.local":
		return fbLocal.App(), r.Host, nil
	case "localhost":
		return fbLocal.App(), "debtstracker.local", nil
	default:
		if strings.HasSuffix(r.Host, ".ngrok.io") {
			return fbLocal.App(), "debtstracker.local", nil
		}
	}

	return nil, "", errors.WithMessage(ErrUnknownHost, r.Host)
}

func getFbAppAndSession(c context.Context, r *http.Request, getSession func(fbApp *fb.App) (*fb.Session, error)) (
	fbApp *fb.App, fbSession *fb.Session, err error,
) {
	log.Debugf(c, "getFbAppAndSession()")
	if fbApp, _, err = GetFbAppAndHost(r); err != nil {
		log.Errorf(c, "getFbAppAndSession() => Failed to get app")
		return nil, nil, err
	}
	if fbSession, err = getSession(fbApp); err != nil {
		log.Errorf(c, "getFbAppAndSession() => Failed to get session")
		return nil, nil, err
	}
	log.Debugf(c, "getFbAppAndSession() => AppId: %v", fbApp.AppId)
	return fbApp, fbSession, err
}

func FbAppAndSessionFromAccessToken(c context.Context, r *http.Request, accessToken string) (*fb.App, *fb.Session, error) {
	return getFbAppAndSession(c, r, func(fbApp *fb.App) (fbSession *fb.Session, err error) {
		fbSession = fbApp.Session(accessToken)
		fbSession.HttpClient = dal.HttpClient(c)
		return
	})
}

func FbAppAndSessionFromSignedRequest(c context.Context, r *http.Request, signedRequest string) (*fb.App, *fb.Session, error) {
	log.Debugf(c, "FbAppAndSessionFromSignedRequest()")
	return getFbAppAndSession(c, r, func(fbApp *fb.App) (fbSession *fb.Session, err error) {
		log.Debugf(c, "FbAppAndSessionFromSignedRequest() => getSession()")
		fbSession, err = fbApp.SessionFromSignedRequest(c, signedRequest, dal.HttpClient(c))
		if err != nil {
			log.Debugf(c, "FbAppAndSessionFromSignedRequest() => getSession(): %v", err.Error())
		}
		return
	})
}
