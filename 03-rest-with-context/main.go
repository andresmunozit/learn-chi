package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

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
		// TODO search route
		r.Get("/", ListArticles)

		r.Route("/{articleID}", func(r chi.Router) {
			r.Use(ArticleCtx)
			r.Get("/", GetArticle)
			r.Put("/", UpdateArticle)
			r.Delete("/", DeleteArticle)
		})
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

func GetArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*Article)

	if err := render.Render(w, r, NewArticleResponse(article)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func UpdateArticle(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value("article").(*Article)

	data := &ArticleRequest{Article: article}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	article = data.Article
	dbUpdateArticle(article.ID, article)

	render.Render(w, r, NewArticleResponse(article))
}

func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	var err error

	article := r.Context().Value("article").(*Article)

	article, err = dbRemoveArticle(article.ID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	render.Render(w, r, NewArticleResponse(article))
}

var ErrBadRequest = &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: "Bad request."}
var InternalServerError = &ErrResponse{HTTPStatusCode: http.StatusInternalServerError, StatusText: "Internal server error."}

// Middlewares
// load an article object from the url parameters
func ArticleCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var article *Article
		var err error

		if articleID := chi.URLParam(r, "articleID"); articleID != "" {
			article, err = dbGetArticle(articleID)
		} else if articleSlug := chi.URLParam(r, "articleSlug"); articleSlug != "" {
			article, err = dbGetArticleBySlug(articleSlug)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// Add article to context
		ctx := context.WithValue(r.Context(), "article", article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Invalid request",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: http.StatusNotFound, StatusText: "Resource not found."}

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Article struct {
	ID     string `json:"id"`
	UserID int64  `json:"user_id"`
	Title  string `json:"title"`
	Slug   string `json:"slug"`
}

type UserPayload struct {
	*User
}

func NewUserPayloadResponse(user *User) *UserPayload {
	return &UserPayload{user}
}

type ArticleRequest struct {
	*Article
	User        *UserPayload `json:"user,omitempty"`
	ProtectedID string       `json:"id"` // override id json to have more control
}

func (a *ArticleRequest) Bind(r *http.Request) error {
	if a.Article == nil {
		return errors.New("missing required Article fields")
	}

	a.ProtectedID = ""
	a.Article.Title = strings.ToLower(a.Article.Title)
	return nil
}

type ArticleResponse struct {
	*Article

	User *UserPayload `json:"user,omitempty"`
}

func (ar *ArticleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewArticleResponse(article *Article) *ArticleResponse {
	resp := &ArticleResponse{Article: article}

	if resp.User == nil {
		if user, _ := dbGetUser(resp.UserID); user != nil {
			resp.User = NewUserPayloadResponse(user)
		}
	}

	return resp
}

func NewArticleListResponse(articles []*Article) []render.Renderer {
	list := []render.Renderer{}
	for _, a := range articles {
		list = append(list, NewArticleResponse(a))
	}
	return list
}

// DB values
var users = []*User{
	{ID: 1, Name: "Andres"},
	{ID: 2, Name: "Mateo"},
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
