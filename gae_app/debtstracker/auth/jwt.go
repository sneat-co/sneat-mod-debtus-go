package auth

import (
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"google.golang.org/appengine"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var secret = []byte("very-secret-abc")

const SECRET_PREFIX = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9."

func IssueToken(userID int64, issuer string, isAdmin bool) string {
	if userID == 0 {
		panic("IssueToken(userID == 0)")
	}
	claims := jws.Claims{}
	claims.SetIssuedAt(time.Now())
	claims.SetSubject(strconv.FormatInt(userID, 10))
	if isAdmin {
		claims.Set("admin", true)
	}

	if issuer != "" {
		if len(issuer) > 100 {
			if len(issuer) <= 1000 {
				panic("IssueToken() => len(issuer) > 20, issuer: " + issuer)
			} else {
				panic("IssueToken() => len(issuer) > 20, issuer[:1000]: " + issuer[:1000])
			}

		}
		claims.SetIssuer(issuer)
	}

	token := jws.NewJWT(claims, crypto.SigningMethodHS256)
	signature, err := token.Serialize(secret)
	if err != nil {
		panic(err.Error())
	}
	return string(signature[len(SECRET_PREFIX):])
}

type AuthInfo struct {
	UserID  int64
	IsAdmin bool
	Issuer  string
}

var ErrNoToken = errors.New("No authorization token")

func Authenticate(w http.ResponseWriter, r *http.Request, required bool) (authInfo AuthInfo, token jwt.JWT, err error) {
	c := appengine.NewContext(r)
	s := r.URL.Query().Get("secret")
	if s == "" {
		if a := r.Header.Get("Authorization"); strings.HasPrefix(a, "Bearer ") {
			s = a[7:]
		}
	}

	defer func() {
		if err != nil && required {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Write([]byte(err.Error()))
		}
	}()

	if s == "" {
		err = ErrNoToken
		return
	}

	if strings.Count(s, ".") == 1 {
		s = SECRET_PREFIX + s
	}

	log.Debugf(appengine.NewContext(r), "JWT token: [%v]", s)

	if token, err = jws.ParseJWT([]byte(s)); err != nil {
		log.Debugf(c, "Tried to parse: [%v]", s)
		return
	}

	claims := token.Claims()
	if sub, ok := claims.Subject(); ok {
		if authInfo.UserID, err = strconv.ParseInt(sub, 10, 64); err == nil {
			authInfo.IsAdmin = claims.Has("admin")
		}
	} else {
		err = errors.New("JWT is missing 'sub' claim.")
		return
	}
	if issuer, ok := claims.Issuer(); ok {
		authInfo.Issuer = issuer
	} else {
		err = errors.New("JWT is missing 'issuer' claim.")
		return
	}
	return
}
