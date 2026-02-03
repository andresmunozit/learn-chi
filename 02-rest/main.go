package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type article struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

var articles = make(map[string]article)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	r.Route("/articles", func(r chi.Router) {
		r.Get("/", getArticles)
		r.Get("/{articleId}", getArticle)
		r.Post("/", createArticle)
	})

	http.ListenAndServe(":3000", r)
}

func getArticles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(articles)
}

func createArticle(w http.ResponseWriter, r *http.Request) {
	var body createArticleBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id, _ := generateId(8)

	a := article{
		Id:   id,
		Text: body.Text,
	}

	articles[id] = a

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func getArticle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "articleId")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a, ok := articles[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a)
}
