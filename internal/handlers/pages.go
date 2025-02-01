package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/custompages"
	"gabe565.com/linx-server/internal/expiry"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/templates"
	"gabe565.com/utils/must"
)

type RespType int

const (
	RespPLAIN RespType = iota
	RespJSON
	RespHTML
	RespAUTO
)

func Index(w http.ResponseWriter, r *http.Request) {
	err := templates.Render("index.html", map[string]any{
		"MaxSize":     config.Default.MaxSize,
		"ExpiryList":  expiry.ListExpirationTimes(),
		"ForceRandom": config.Default.ForceRandomFilename,
	}, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Paste(w http.ResponseWriter, r *http.Request) {
	err := templates.Render("paste.html", map[string]any{
		"ExpiryList":  expiry.ListExpirationTimes(),
		"ForceRandom": config.Default.ForceRandomFilename,
	}, r, w)
	if err != nil {
		Oops(w, r, RespHTML, "")
	}
}

func APIDoc(w http.ResponseWriter, r *http.Request) {
	err := templates.Render("API.html", map[string]any{
		"SiteURL":     must.Must2(headers.GetSiteURL(r)),
		"ForceRandom": config.Default.ForceRandomFilename,
	}, r, w)
	if err != nil {
		Oops(w, r, RespHTML, "")
	}
}

func MakeCustomPage(fileName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := templates.Render("custom_page.html", map[string]any{
			"SiteURL":     must.Must2(headers.GetSiteURL(r)),
			"ForceRandom": config.Default.ForceRandomFilename,
			"Contents":    template.HTML(custompages.CustomPages[fileName]),
			"FileName":    fileName,
			"PageName":    custompages.Names[fileName],
		}, r, w)
		if err != nil {
			Oops(w, r, RespHTML, "")
		}
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	err := templates.Render("404.html", nil, r, w)
	if err != nil {
		Oops(w, r, RespHTML, "")
	}
}

func Oops(w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	if msg == "" {
		msg = "Oops! Something went wrong..."
	}

	if rt == RespHTML {
		w.WriteHeader(500)
		templates.Render("oops.html", map[string]any{"Msg": msg}, r, w)
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
			Oops(w, r, RespJSON, msg)
		} else {
			Oops(w, r, RespHTML, msg)
		}
	}
}

func BadRequest(w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	if rt == RespHTML {
		w.WriteHeader(http.StatusBadRequest)
		err := templates.Render("400.html", map[string]any{"Msg": msg}, r, w)
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
			BadRequest(w, r, RespJSON, msg)
		} else {
			BadRequest(w, r, RespHTML, msg)
		}
	}
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
	err := templates.Render("401.html", nil, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
