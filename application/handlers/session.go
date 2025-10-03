package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"git.miem.hse.ru/kg25-26/aisavelev.git/application/models"
)

var (
	sessions    = make(map[string]*models.Session)
	sessionsMux sync.RWMutex
	targetColor models.ColorRGBA
)

func GetSession(r *http.Request) *models.Session {
	sessionID := getSessionID(r)

	sessionsMux.Lock()
	defer sessionsMux.Unlock()

	if session, exists := sessions[sessionID]; exists {
		return session
	}

	// Создание новой сессии
	session := &models.Session{
		ID:        sessionID,
		Images:    []models.ImageAnalysis{},
		CreatedAt: time.Now(),
	}
	sessions[sessionID] = session
	return session
}

func getSessionID(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Id сессии
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
	})
}

func CleanupSessions() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sessionsMux.Lock()
		now := time.Now()
		for id, session := range sessions {
			if now.Sub(session.CreatedAt) > 24*time.Hour {
				// Автоматическая очистка сессии
				for _, img := range session.Images {
					os.Remove(filepath.Join("static", "uploads", session.ID, img.FileName))
				}
				os.Remove(filepath.Join("static", "uploads", session.ID))
				delete(sessions, id)
			}
		}
		sessionsMux.Unlock()
	}
}

func GetTargetColor() models.ColorRGBA {
	return targetColor
}

func SetTargetColor(color models.ColorRGBA) {
	targetColor = color
}
