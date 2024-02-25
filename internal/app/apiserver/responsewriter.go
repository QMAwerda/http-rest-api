package apiserver

import "net/http"

type responseWriter struct {
	http.ResponseWriter     // все эти методы будут доступны внутри нашей структуры, их не нужно реализовывать
	code                int // статус код нашего ответа
}

// Переопределим метод для записи кодов ответа
func (w *responseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
