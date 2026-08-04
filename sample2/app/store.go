package app

import (
	"errors"
	"strings"
	"sync"
)

// User data model.
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Article data model.
type Article struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Title  string `json:"title"`
	Slug   string `json:"slug"`
}

// ArticleRequest is the create/update payload.
type ArticleRequest struct {
	Title string `json:"title"`
}

// ---------- mock database ----------

var (
	mu       sync.Mutex
	articles []Article
	users    []User
	nextID   int64 = 1
)

func init() {
	users = []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
	articles = []Article{
		{ID: 1, UserID: 1, Title: "First Post", Slug: "first-post"},
		{ID: 2, UserID: 1, Title: "Second Post", Slug: "second-post"},
		{ID: 3, UserID: 2, Title: "Hello World", Slug: "hello-world"},
		{ID: 4, UserID: 2, Title: "Go Is Great", Slug: "go-is-great"},
		{ID: 5, UserID: 1, Title: "REST APIs", Slug: "rest-apis"},
	}
	nextID = 6
}

var (
	ErrNotFound = errors.New("not found")
)

func ResetStore() {
	mu.Lock()
	defer mu.Unlock()
	users = []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
	articles = []Article{
		{ID: 1, UserID: 1, Title: "First Post", Slug: "first-post"},
		{ID: 2, UserID: 1, Title: "Second Post", Slug: "second-post"},
		{ID: 3, UserID: 2, Title: "Hello World", Slug: "hello-world"},
		{ID: 4, UserID: 2, Title: "Go Is Great", Slug: "go-is-great"},
		{ID: 5, UserID: 1, Title: "REST APIs", Slug: "rest-apis"},
	}
	nextID = 6
}

func DBGetArticle(id int64) (Article, error) {
	mu.Lock()
	defer mu.Unlock()
	for _, a := range articles {
		if a.ID == id {
			return a, nil
		}
	}
	return Article{}, ErrNotFound
}

func DBGetArticleBySlug(slug string) (Article, error) {
	mu.Lock()
	defer mu.Unlock()
	for _, a := range articles {
		if a.Slug == slug {
			return a, nil
		}
	}
	return Article{}, ErrNotFound
}

func DBListArticles() []Article {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Article, len(articles))
	copy(out, articles)
	return out
}

func DBSearchArticles(q string) []Article {
	mu.Lock()
	defer mu.Unlock()
	result := make([]Article, 0)
	lower := strings.ToLower(q)
	for _, a := range articles {
		if strings.Contains(strings.ToLower(a.Title), lower) || strings.Contains(strings.ToLower(a.Slug), lower) {
			result = append(result, a)
		}
	}
	return result
}

func DBCreateArticle(title string, userID int64) Article {
	mu.Lock()
	defer mu.Unlock()
	a := Article{
		ID:     nextID,
		UserID: userID,
		Title:  title,
		Slug:   strings.ToLower(strings.ReplaceAll(title, " ", "-")),
	}
	nextID++
	articles = append(articles, a)
	return a
}

func DBUpdateArticle(id int64, title string) (Article, error) {
	mu.Lock()
	defer mu.Unlock()
	for i, a := range articles {
		if a.ID == id {
			articles[i].Title = title
			articles[i].Slug = strings.ToLower(strings.ReplaceAll(title, " ", "-"))
			return articles[i], nil
		}
	}
	return Article{}, ErrNotFound
}

func DBDeleteArticle(id int64) error {
	mu.Lock()
	defer mu.Unlock()
	for i, a := range articles {
		if a.ID == id {
			articles = append(articles[:i], articles[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func DBGetUser(id int64) (User, error) {
	mu.Lock()
	defer mu.Unlock()
	for _, u := range users {
		if u.ID == id {
			return u, nil
		}
	}
	return User{}, ErrNotFound
}
