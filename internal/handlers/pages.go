package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gabe565.com/linx-server/internal/template"
)

type RespType int

const (
	RespAUTO RespType = iota
	RespHTML
	RespJSON
	RespPLAIN
)

func Error(w http.ResponseWriter, r *http.Request, status int) {
	ErrorType(w, r, RespAUTO, status, "")
}

func ErrorMsg(w http.ResponseWriter, r *http.Request, status int, msg string) {
	ErrorType(w, r, RespAUTO, status, msg)
}

func ErrorType(w http.ResponseWriter, r *http.Request, rt RespType, status int, msg string) {
	if msg == "" {
		msg = http.StatusText(status)
	}

	w.Header().Set("Cache-Control", "no-store")

	switch rt {
	case RespAUTO:
		switch {
		case strings.EqualFold("application/json", r.Header.Get("Accept")):
			ErrorType(w, r, RespJSON, status, msg)
			return
		case IsDirectUA(r):
			ErrorType(w, r, RespPLAIN, status, msg)
			return
		default:
			ErrorType(w, r, RespHTML, status, msg)
		}
	case RespHTML:
		ServeAsset(w, r, status, template.WithTitle(http.StatusText(status)))
	case RespJSON:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
	case RespPLAIN:
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, msg)
	}
}
