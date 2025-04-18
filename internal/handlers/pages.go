package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/custompages"
	"gabe565.com/linx-server/internal/expiry"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/templates"
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
		"MaxSize":     int(config.Default.MaxSize),
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
		return
	}
}

func APIDoc(w http.ResponseWriter, r *http.Request) {
	err := templates.Render("API.html", map[string]any{
		"SiteURL":     headers.GetSiteURL(r).String(),
		"ForceRandom": config.Default.ForceRandomFilename,
	}, r, w)
	if err != nil {
		Oops(w, r, RespHTML, "")
		return
	}
}

func MakeCustomPage(fileName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := templates.Render("custom_page.html", map[string]any{
			"SiteURL":     headers.GetSiteURL(r).String(),
			"ForceRandom": config.Default.ForceRandomFilename,
			"Contents":    template.HTML(custompages.CustomPages[fileName]), //nolint:gosec
			"FileName":    fileName,
			"PageName":    custompages.Names[fileName],
		}, r, w)
		if err != nil {
			Oops(w, r, RespHTML, "")
			return
		}
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	err := templates.Render("404.html", nil, r, w)
	if err != nil {
		Oops(w, r, RespHTML, "")
		return
	}
}

func Oops(w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	if msg == "" {
		msg = "Oops! Something went wrong..."
	}

	const name = "error.html"

	switch rt {
	case RespHTML:
		w.WriteHeader(http.StatusInternalServerError)
		if err := templates.Render(name, map[string]any{"Msg": msg}, r, w); err != nil {
			slog.Error("Failed to render template", "template", name, "error", err)
		}
		return
	case RespPLAIN:
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "%s", msg)
		return
	case RespJSON:
		js, _ := json.Marshal(map[string]string{
			"error": msg,
		})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(js)
		return
	case RespAUTO:
		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			Oops(w, r, RespJSON, msg)
			return
		}
		Oops(w, r, RespHTML, msg)
		return
	}
}

func BadRequest(w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	switch rt {
	case RespHTML:
		w.WriteHeader(http.StatusBadRequest)
		err := templates.Render("error.html", map[string]any{"Title": "400 Bad Request", "Msg": msg}, r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	case RespPLAIN:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", msg)
		return
	case RespJSON:
		js, _ := json.Marshal(map[string]string{
			"error": msg,
		})
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(js)
		return
	case RespAUTO:
		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			BadRequest(w, r, RespJSON, msg)
			return
		}
		BadRequest(w, r, RespHTML, msg)
		return
	}
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	err := templates.Render("error.html", map[string]any{"Title": "401 Unauthorized"}, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
