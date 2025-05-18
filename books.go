package main

import "time"

type Book struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewBook(id int, name, author string, createdAt time.Time) *Book {
	return &Book{
		ID:        id,
		Name:      name,
		Author:    author,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}
