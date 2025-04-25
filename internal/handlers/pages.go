package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RespType int

const (
	RespPLAIN RespType = iota
	RespJSON
	RespHTML
	RespAUTO
)

// func MakeCustomPage(fileName string) func(w http.ResponseWriter, r *http.Request) {
//	return func(w http.ResponseWriter, r *http.Request) {
//		err := templates.Render("custom_page.html", map[string]any{
//			"SiteURL":     headers.GetSiteURL(r).String(),
//			"ForceRandom": config.Default.ForceRandomFilename,
//			"Contents":    template.HTML(custompages.CustomPages[fileName]), //nolint:gosec
//			"FileName":    fileName,
//			"PageName":    custompages.Names[fileName],
//		}, r, w)
//		if err != nil {
//			Oops(w, r, RespHTML, "")
//			return
//		}
//	}
//}

func NotFound(w http.ResponseWriter, r *http.Request) {
	AssetHandler(w, r)
}

func Oops(w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	if msg == "" {
		msg = "Oops! Something went wrong..."
	}

	switch rt {
	case RespHTML:
		AssetHandler(w, r)
		return
	case RespPLAIN:
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "%s", msg)
		return
	case RespJSON:
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
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
		AssetHandler(w, r)
		return
	case RespPLAIN:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", msg)
		return
	case RespJSON:
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
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
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: http.StatusText(http.StatusUnauthorized)})
		return
	}
	AssetHandler(w, r)
}
