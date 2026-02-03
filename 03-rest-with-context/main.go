package main

import (
	"errors"
	"fmt"
	"math/rand"
)

func main() {

}

// Types
type User struct {
	ID   int64
	Name string
}

type Article struct {
	ID     string
	UserID int64
	Title  string
	Slug   string
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
