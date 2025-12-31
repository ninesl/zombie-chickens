package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const SessionCookieName = "zc_session"

type contextKey string

const SessionIDKey contextKey = "sessionID"

// Session middleware ensures every request has a session ID
func Session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessionID string

		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			// Generate new session ID
			sessionID = generateSessionID()
			http.SetCookie(w, &http.Cookie{
				Name:     SessionCookieName,
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   86400 * 7, // 7 days
			})
		} else {
			sessionID = cookie.Value
		}

		ctx := context.WithValue(r.Context(), SessionIDKey, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSessionID retrieves session ID from context
func GetSessionID(ctx context.Context) string {
	sessionID, _ := ctx.Value(SessionIDKey).(string)
	return sessionID
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
