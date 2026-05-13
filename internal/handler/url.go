// internal/handler/url.go
package handler

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/OlegSereb/go-musthave-shortener-tpl.git/internal/service"
)

// URLHandler обрабатывает HTTP-запросы, связанные с URL
type URLHandler struct {
	service *service.URLService
}

// NewURLHandler создаёт новый экземпляр обработчика
func NewURLHandler(svc *service.URLService) *URLHandler {
	return &URLHandler{service: svc}
}

// ShortenURL обрабатывает POST / — создаёт короткую ссылку
// Ожидает: тело запроса с оригинальным URL в формате text/plain
// Возвращает: 201 + короткая ссылка, или 400 при ошибке
func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	// Логируем входящий запрос для отладки
	log.Printf("POST / : попытка сократить URL")

	// Проверяем метод запроса (дополнительная защита)
	if r.Method != http.MethodPost {
		log.Printf("  ❌ неверный метод: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("  ❌ ошибка чтения тела: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Извлекаем и очищаем URL
	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		log.Printf("  ❌ пустой URL")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("  📥 оригинальный URL: %s", originalURL)

	// Вызываем сервис для создания короткой ссылки
	shortURL, err := h.service.Shorten(originalURL)
	if err != nil {
		log.Printf("  ❌ ошибка в сервисе: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("  ✅ создана короткая ссылка: %s", shortURL)

	// Формируем успешный ответ
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated) // 201
	// Отправляем короткую ссылку в теле ответа (без лишних символов!)
	_, _ = w.Write([]byte(shortURL))
}

// RedirectURL обрабатывает GET /{id} — перенаправляет на оригинальный URL
// Ожидает: короткий идентификатор в пути запроса
// Возвращает: 307 + заголовок Location, или 400 при ошибке
func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	// Логируем входящий запрос
	log.Printf("GET %s : поиск оригинального URL", r.URL.Path)

	// Проверяем метод — разрешаем только GET
	if r.Method != http.MethodGet {
		log.Printf("  ❌ неверный метод: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Извлекаем короткий ID из пути: /EwHXdJfB → EwHXdJfB
	shortID := strings.TrimPrefix(r.URL.Path, "/")

	// Проверяем, что ID не пустой
	if shortID == "" {
		log.Printf("  ❌ пустой идентификатор")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("  🔍 поиск по ID: %s", shortID)

	// Ищем оригинальный URL в сервисе
	originalURL, exists := h.service.GetOriginal(shortID)
	if !exists {
		log.Printf("  ❌ ID не найден: %s", shortID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("  ✅ найдено: %s → %s", shortID, originalURL)

	// Выполняем временное перенаправление (код 307)
	// 307 сохраняет метод и тело запроса при редиректе (в отличие от 302)
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
