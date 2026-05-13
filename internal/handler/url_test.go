// internal/handler/url_test.go
package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OlegSereb/go-musthave-shortener-tpl.git/internal/service"
)

// TestShortenURL тестирует обработчик POST /
func TestShortenURL(t *testing.T) {
	svc := service.NewURLService("http://localhost:8080")
	handler := NewURLHandler(svc)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "валидный URL",
			requestBody:    "https://practicum.yandex.ru/",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/",
		},
		{
			name:           "пустое тело запроса",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "только пробелы",
			requestBody:    "   \n\t  ",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "text/plain")
			rr := httptest.NewRecorder()

			handler.ShortenURL(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("ShortenURL() status = %d, want %d", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				body := rr.Body.String()
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("ShortenURL() body = %q, want to contain %q", body, tt.expectedBody)
				}
				// Важно для автотестов: без \n в конце
				if strings.HasSuffix(body, "\n") {
					t.Errorf("ShortenURL() body has trailing newline: %q", body)
				}
			}

			if contentType := rr.Header().Get("Content-Type"); contentType != "text/plain" {
				t.Errorf("ShortenURL() Content-Type = %q, want %q", contentType, "text/plain")
			}
		})
	}
}

// TestRedirectURL_WithKnownID тестирует редирект для существующего ID
func TestRedirectURL_WithKnownID(t *testing.T) {
	svc := service.NewURLService("http://localhost:8080")
	handler := NewURLHandler(svc)

	// Создаём тестовые данные через публичный API сервиса
	original := "https://test.example.com/path"
	shortURL, err := svc.Shorten(original)
	if err != nil {
		t.Fatalf("failed to create test data: %v", err)
	}

	// Извлекаем ID из короткого URL
	parts := strings.Split(shortURL, "/")
	shortID := parts[len(parts)-1]

	// Создаём запрос на редирект
	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	rr := httptest.NewRecorder()

	handler.RedirectURL(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}
	if loc := rr.Header().Get("Location"); loc != original {
		t.Errorf("Location = %q, want %q", loc, original)
	}
}

// TestRedirectURL_EdgeCases тестирует граничные случаи
func TestRedirectURL_EdgeCases(t *testing.T) {
	svc := service.NewURLService("http://localhost:8080")
	handler := NewURLHandler(svc)

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "ID не найден",
			path:           "/NonExistent",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "пустой путь",
			path:           "/",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "неверный метод (POST вместо GET)",
			path:           "/SomeID",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			handler.RedirectURL(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("RedirectURL() status = %d, want %d", status, tt.expectedStatus)
			}
		})
	}
}
