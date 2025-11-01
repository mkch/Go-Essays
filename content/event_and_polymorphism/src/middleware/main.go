package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

type Handler func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)

type Middleware http.HandlerFunc

func (m *Middleware) Use(handler Handler) {
	old := *m
	*m = func(w http.ResponseWriter, r *http.Request) {
		next := func(w http.ResponseWriter, r *http.Request) {
			if old != nil {
				old(w, r)
			}
		}
		handler(w, r, next)
	}
}

func main() {
	// Middleware to handle root path "/"
	var rootMiddleware Middleware = func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello!")
	}

	// Add Authorization middleware.
	rootMiddleware.Use(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if r.FormValue("user") != "admin" {
			code := http.StatusUnauthorized
			http.Error(w, http.StatusText(code), code)
			return
		}
		next(w, r)
	})

	// Add log middleware
	rootMiddleware.Use(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		start := time.Now()
		next(w, r)
		duration := time.Since(start)
		log.Printf("Path: %v Duration: %v", r.URL.Path, duration)
	})

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		rootMiddleware(w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
