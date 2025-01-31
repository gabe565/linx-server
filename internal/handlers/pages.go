package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/andreimarcu/linx-server/internal/custompages"
	"github.com/andreimarcu/linx-server/internal/expiry"
	"github.com/andreimarcu/linx-server/internal/headers"
	"github.com/andreimarcu/linx-server/internal/templates"
	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
)

type RespType int

const (
	RespPLAIN RespType = iota
	RespJSON
	RespHTML
	RespAUTO
)

func Index(c web.C, w http.ResponseWriter, r *http.Request) {
	err := templates.Render(config.Templates["index.html"], pongo2.Context{
		"maxsize":     config.Default.MaxSize,
		"expirylist":  expiry.ListExpirationTimes(),
		"forcerandom": config.Default.ForceRandomFilename,
	}, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Paste(c web.C, w http.ResponseWriter, r *http.Request) {
	err := templates.Render(config.Templates["paste.html"], pongo2.Context{
		"expirylist":  expiry.ListExpirationTimes(),
		"forcerandom": config.Default.ForceRandomFilename,
	}, r, w)
	if err != nil {
		Oops(c, w, r, RespHTML, "")
	}
}

func APIDoc(c web.C, w http.ResponseWriter, r *http.Request) {
	err := templates.Render(config.Templates["API.html"], pongo2.Context{
		"siteurl":     headers.GetSiteURL(r),
		"forcerandom": config.Default.ForceRandomFilename,
	}, r, w)
	if err != nil {
		Oops(c, w, r, RespHTML, "")
	}
}

func MakeCustomPage(fileName string) func(c web.C, w http.ResponseWriter, r *http.Request) {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		err := templates.Render(config.Templates["custom_page.html"], pongo2.Context{
			"siteurl":     headers.GetSiteURL(r),
			"forcerandom": config.Default.ForceRandomFilename,
			"contents":    custompages.CustomPages[fileName],
			"filename":    fileName,
			"pagename":    custompages.Names[fileName],
		}, r, w)
		if err != nil {
			Oops(c, w, r, RespHTML, "")
		}
	}
}

func NotFound(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	err := templates.Render(config.Templates["404.html"], pongo2.Context{}, r, w)
	if err != nil {
		Oops(c, w, r, RespHTML, "")
	}
}

func Oops(c web.C, w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	if msg == "" {
		msg = "Oops! Something went wrong..."
	}

	if rt == RespHTML {
		w.WriteHeader(500)
		templates.Render(config.Templates["oops.html"], pongo2.Context{"msg": msg}, r, w)
		return
	} else if rt == RespPLAIN {
		w.WriteHeader(500)
		fmt.Fprintf(w, "%s", msg)
		return
	} else if rt == RespJSON {
		js, _ := json.Marshal(map[string]string{
			"error": msg,
		})

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(500)
		w.Write(js)
		return
	} else if rt == RespAUTO {
		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			Oops(c, w, r, RespJSON, msg)
		} else {
			Oops(c, w, r, RespHTML, msg)
		}
	}
}

func BadRequest(c web.C, w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	if rt == RespHTML {
		w.WriteHeader(http.StatusBadRequest)
		err := templates.Render(config.Templates["400.html"], pongo2.Context{"msg": msg}, r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if rt == RespPLAIN {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", msg)
		return
	} else if rt == RespJSON {
		js, _ := json.Marshal(map[string]string{
			"error": msg,
		})

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(js)
		return
	} else if rt == RespAUTO {
		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			BadRequest(c, w, r, RespJSON, msg)
		} else {
			BadRequest(c, w, r, RespHTML, msg)
		}
	}
}

func Unauthorized(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
	err := templates.Render(config.Templates["401.html"], pongo2.Context{}, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
