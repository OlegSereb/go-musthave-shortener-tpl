// internal/service/url.go
package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/OlegSereb/go-musthave-shortener-tpl.git/internal/model"
)

type URLService struct {
	mu      sync.RWMutex
	store   map[string]model.URL // short_id -> URL
	baseURL string               // например, "http://localhost:8080"
}

func NewURLService(baseURL string) *URLService {
	return &URLService{
		store:   make(map[string]model.URL),
		baseURL: baseURL,
	}
}

// generateShortID создаёт случайную строку из 8 символов
func (s *URLService) generateShortID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// base64 URL-encoding: безопасно для URL, без спецсимволов
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Shorten создаёт короткую ссылку и сохраняет соответствие
func (s *URLService) Shorten(originalURL string) (string, error) {
	if originalURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	shortID, err := s.generateShortID()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[shortID] = model.URL{
		ShortID:     shortID,
		OriginalURL: originalURL,
	}

	return s.baseURL + "/" + shortID, nil
}

// GetOriginal возвращает оригинальный URL по короткому ID
func (s *URLService) GetOriginal(shortID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exists := s.store[shortID]
	if !exists {
		return "", false
	}
	return url.OriginalURL, true
}
