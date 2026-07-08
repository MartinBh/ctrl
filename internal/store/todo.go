package store

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrEmptyTodoTitle = errors.New("todo title cannot be empty")
	ErrTodoNotFound   = errors.New("todo not found")
)

type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type TodoStore struct {
	path string
}

func NewTodoStore(path string) *TodoStore {
	return &TodoStore{path: path}
}

func (s *TodoStore) Load() ([]Todo, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Todo{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []Todo{}, nil
	}

	var todos []Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}

	return todos, nil
}

func (s *TodoStore) Add(title string) ([]Todo, Todo, error) {
	title, err := normalizeTodoTitle(title)
	if err != nil {
		return nil, Todo{}, err
	}

	todos, err := s.Load()
	if err != nil {
		return nil, Todo{}, err
	}

	now := time.Now().UTC()
	todo := Todo{
		ID:        newTodoID(now),
		Title:     title,
		CreatedAt: now,
	}
	todos = append(todos, todo)

	if err := s.Save(todos); err != nil {
		return nil, Todo{}, err
	}

	return todos, todo, nil
}

func (s *TodoStore) UpdateTitle(id string, title string) ([]Todo, error) {
	title, err := normalizeTodoTitle(title)
	if err != nil {
		return nil, err
	}

	return s.update(id, func(todo *Todo) {
		todo.Title = title
	})
}

func (s *TodoStore) Toggle(id string) ([]Todo, error) {
	return s.update(id, func(todo *Todo) {
		todo.Done = !todo.Done
	})
}

func (s *TodoStore) Delete(id string) ([]Todo, error) {
	todos, err := s.Load()
	if err != nil {
		return nil, err
	}

	for index, todo := range todos {
		if todo.ID == id {
			todos = append(todos[:index], todos[index+1:]...)
			if err := s.Save(todos); err != nil {
				return nil, err
			}
			return todos, nil
		}
	}

	return nil, ErrTodoNotFound
}

func (s *TodoStore) Save(todos []Todo) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, append(data, '\n'), 0o644)
}

func (s *TodoStore) update(id string, apply func(*Todo)) ([]Todo, error) {
	todos, err := s.Load()
	if err != nil {
		return nil, err
	}

	for index := range todos {
		if todos[index].ID == id {
			apply(&todos[index])
			if err := s.Save(todos); err != nil {
				return nil, err
			}
			return todos, nil
		}
	}

	return nil, ErrTodoNotFound
}

func normalizeTodoTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ErrEmptyTodoTitle
	}

	return title, nil
}

func newTodoID(now time.Time) string {
	var suffix [4]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return fmt.Sprintf("%d", now.UnixNano())
	}

	return fmt.Sprintf("%d-%x", now.UnixNano(), suffix)
}
