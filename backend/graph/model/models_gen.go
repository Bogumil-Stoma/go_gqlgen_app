// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Mutation struct {
}

type Query struct {
}

type Translation struct {
	WordID        string `json:"wordID"`
	TranslationID string `json:"translationID"`
}

type Word struct {
	ID       string `json:"id"`
	Word     string `json:"word"`
	Language string `json:"language"`
}
