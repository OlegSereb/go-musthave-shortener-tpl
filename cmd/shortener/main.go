// cmd/shortener/main.go
package main

import (
	"log"
	"net/http"

	"github.com/OlegSereb/go-musthave-shortener-tpl.git/internal/handler"
	"github.com/OlegSereb/go-musthave-shortener-tpl.git/internal/service"
)

func main() {
	// Инициализируем сервис
	urlService := service.NewURLService("http://localhost:8080")
	urlHandler := handler.NewURLHandler(urlService)

	// Настраиваем маршруты
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == http.MethodPost {
			urlHandler.ShortenURL(w, r)
		} else if r.URL.Path != "/" && r.Method == http.MethodGet {
			urlHandler.RedirectURL(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	// Запускаем сервер
	log.Println("Starting server at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
