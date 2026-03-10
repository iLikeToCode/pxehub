package httpserver

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"pxehub/internal/db"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
)

var (
	sessions   = make(map[string]string)
	sessionsMu sync.Mutex
)

func generateSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func createSession(w http.ResponseWriter, r *http.Request, username string) {
	sessionID := generateSessionID()

	sessionsMu.Lock()
	sessions[sessionID] = username
	sessionsMu.Unlock()

	secure := false
	if r.TLS != nil {
		secure = true
	} else if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		secure = true
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func getSessionUsername(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", false
	}

	sessionsMu.Lock()
	username, ok := sessions[cookie.Value]
	sessionsMu.Unlock()

	return username, ok
}

func (h *HttpServer) LoginPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := db.GetUserByUsername(username, h.Database)
	if err != nil || db.CheckPassword(user.Password, password) != nil {
		http.Redirect(w, r, "/login?error=Invalid+credentials", http.StatusSeeOther)
		return
	}

	createSession(w, r, username)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
