package api

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	//"encoding/json"
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app"
	"github.com/strongo/db"
	"github.com/strongo/log"
)

func getEnvironment(r *http.Request) strongo.Environment {
	switch r.Host {
	case "debtstracker.io":
		return strongo.EnvProduction
	case "debtstracker-dev1.appspot.com":
		return strongo.EnvDevTest
	case "debtstracker.local":
		return strongo.EnvLocal
	case "localhost":
		return strongo.EnvLocal
	default:
		panic("Unknonwn host: " + r.Host)
	}
}

func getID(c context.Context, w http.ResponseWriter, r *http.Request, idParamName string) int64 {
	q := r.URL.Query()
	if idParamName == "" {
		panic("idParamName is not specified")
	}
	idParamVal := q.Get(idParamName)
	if id, err := strconv.ParseInt(idParamVal, 10, 64); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to decode to int64: '" + idParamVal + "'"))
		return 0
	} else {
		log.Infof(c, "Int ID: %v", id)
		return id
	}
}

func hasError(c context.Context, w http.ResponseWriter, err error, entity string, id int64, notFoundStatus int) bool {
	switch {
	case err == nil:
		return false
	case db.IsNotFound(err):
		if notFoundStatus == 0 {
			notFoundStatus = http.StatusNotFound
		}
		w.WriteHeader(notFoundStatus)
		m := fmt.Sprintf("Entity %v not found by id: %d", entity, id)
		log.Infof(c, m)
		w.Write([]byte(m))
	default:
		err = errors.Wrapf(err, "Failed to get entity %v by id=%v", entity, id)
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	return true
}

func jsonToResponse(c context.Context, w http.ResponseWriter, v interface{}) {
	header := w.Header()
	if buffer, err := ffjson.Marshal(v); err != nil {
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		header.Add("Access-Control-Allow-Origin", "*")
		log.Debugf(c, "w.Header(): %v", header)
		w.Write([]byte(err.Error()))
	} else {
		markResponseAsJson(header)
		log.Debugf(c, "w.Header(): %v", header)
		_, err := w.Write(buffer)
		ffjson.Pool(buffer)
		if err != nil {
			InternalError(c, w, err)
		}
	}
}

func ErrorAsJson(c context.Context, w http.ResponseWriter, status int, err error) {
	if status == 0 {
		panic("status == 0")
	}
	if status == http.StatusInternalServerError {
		log.Errorf(c, "Error: %v", err.Error())
	} else {
		log.Infof(c, "Error: %v", err.Error())
	}
	w.WriteHeader(status)
	jsonToResponse(c, w, map[string]string{"error": err.Error()})
}

func markResponseAsJson(header http.Header) {
	header.Add("Content-Type", "application/json")
	header.Add("Access-Control-Allow-Origin", "*")
}
