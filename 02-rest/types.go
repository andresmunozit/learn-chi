package main

type createArticleBody struct {
	Text string `json:"text"`
}

type updateArticleBody struct {
	Text string `json:"text"`
}
