package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.AllowContentType("application/json"))

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
		}{"ok"})
	})

	r.Route("/articles", func(r chi.Router) {
		// r.Use(ArticleCtx)
		r.Get("/", ListArticles)
	})

	http.ListenAndServe(":3000", r)
}

// Handlers
func ListArticles(w http.ResponseWriter, r *http.Request) {
	if err := render.RenderList(w, r, NewArticleListResponse(articles)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// Types
type ErrResponse struct {
	Err            error  `json:"-"`
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	AppCode        int64  `json:"code,omitempty"`
	ErrorText      string `json:"error,omitempty"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

// Middleware
// func ArticleCtx(h http.Handler) http.Handler {

// }

type User struct {
	ID   int64
	Name string
}

type Article struct {
	ID     string `json:"id"`
	UserID int64  `json:"user_id"`
	Title  string `json:"title"`
	Slug   string `json:"slug"`
}

type ArticleResponse struct {
	*Article
}

func (ar *ArticleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewArticleListResponse(articles []*Article) []render.Renderer {
	var renderers []render.Renderer
	for _, a := range articles {
		renderers = append(renderers, &ArticleResponse{a})
	}
	return renderers
}

// DB values
var users = []*User{
	{ID: 1, Name: "Andres"},
	{ID: 1, Name: "Mateo"},
}

var articles = []*Article{
	{ID: "1", UserID: 1, Title: "Lorem", Slug: "lorem"},
	{ID: "2", UserID: 1, Title: "Ipsum", Slug: "impsum"},
	{ID: "3", UserID: 1, Title: "Dolor", Slug: "dolor"},
	{ID: "4", UserID: 2, Title: "Sit", Slug: "sit"},
	{ID: "5", UserID: 2, Title: "Amet", Slug: "amet"},
}

// DB functions
func dbNewArticle(article *Article) (string, error) {
	article.ID = fmt.Sprintf("%d", rand.Intn(100)+10)
	articles = append(articles, article)
	return article.ID, nil
}

func dbGetArticle(id string) (*Article, error) {
	for _, a := range articles {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("article not found.")
}

func dbGetArticleBySlug(slug string) (*Article, error) {
	for _, a := range articles {
		if a.Slug == slug {
			return a, nil
		}
	}
	return nil, errors.New("article not found.")
}

func dbUpdateArticle(id string, article *Article) (*Article, error) {
	for i, a := range articles {
		if a.ID == id {
			articles[i] = article
			return article, nil
		}
	}
	return nil, errors.New("article not found.")
}

func dbRemoveArticle(id string) (*Article, error) {
	for i, a := range articles {
		if a.ID == id {
			articles = append(articles[:i], articles[i+1:]...)
			return a, nil
		}
	}
	return nil, errors.New("article not found.")
}

func dbGetUser(id int64) (*User, error) {
	for _, u := range users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("user not found.")
}
