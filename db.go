package main

import (
	"errors"
	"slices"
	"sync"
	"time"
)

type DB struct {
	mutex      sync.RWMutex
	books      []*Book
	lastBookID int
	index      map[int]int
}

var (
	once sync.Once
)

func NewDB() *DB {
	var db *DB
	once.Do(func() {
		if db == nil {
			db = &DB{
				index: make(map[int]int),
			}
		}
	})

	return db
}

func (db *DB) newBookID() int {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id := db.lastBookID + 1
	db.lastBookID = id

	return id
}

func (db *DB) AddBook(name, author string) *Book {
	book := NewBook(db.newBookID(), name, author, time.Now())

	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.index[book.ID] = len(db.books)
	db.books = append(db.books, book)

	return book
}

func (db *DB) GetBooks() []*Book {
	return db.books
}

var (
	ErrBookNotFound = errors.New("error book not found")
)

func (db *DB) GetBook(id int) (*Book, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	index, ok := db.index[id]
	if !ok {
		return nil, ErrBookNotFound
	}

	return db.books[index], nil
}

func (db *DB) RemoveBook(id int) error {
	db.mutex.RLock()
	index, ok := db.index[id]
	db.mutex.RUnlock()
	if !ok {
		return ErrBookNotFound
	}

	db.mutex.Lock()
	db.mutex.Unlock()
	db.books = slices.Delete(db.books, index, index+1)

	delete(db.index, id)

	return nil
}

func (db *DB) UpdateBook(id int, name, author string) (*Book, error) {
	db.mutex.RLock()
	index, ok := db.index[id]
	if !ok {
		db.mutex.RUnlock()
		return nil, ErrBookNotFound
	}

	book := db.books[index]
	db.mutex.RUnlock()

	book.Name = name
	book.Author = author
	book.UpdatedAt = time.Now()

	db.mutex.Lock()
	db.mutex.Unlock()
	db.books[index] = book

	return book, nil
}
