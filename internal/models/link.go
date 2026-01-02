package models

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"
)

// Link представляет ссылку в системе
type Link struct {
	ID          int32     `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortName   string    `json:"short_name"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

// CreateLinkRequest запрос на создание ссылки
type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name,omitempty" binding:"omitempty,min=3,max=32"`
}

// UpdateLinkRequest запрос на обновление ссылки
type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name" binding:"required,min=3,max=32"`
}

// GenerateShortName генерирует уникальное короткое имя
func GenerateShortName() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Кодируем в base64 и берем первые 8 символов
	encoded := base64.URLEncoding.EncodeToString(b)
	// Убираем специальные символы и делаем короче
	shortName := strings.ReplaceAll(encoded, "-", "")
	shortName = strings.ReplaceAll(shortName, "_", "")

	if len(shortName) > 8 {
		shortName = shortName[:8]
	}

	return shortName, nil
}
