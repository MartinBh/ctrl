package todo

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/store"
	"github.com/martinbhatta/ctrl/internal/theme"
)

type Panel struct {
	list  *tview.List
	todos []store.Todo
}

func NewPanel() *Panel {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	theme.Box(list.Box, "TODO")

	return &Panel{list: list}
}

func (p *Panel) Primitive() tview.Primitive {
	return p.list
}

func (p *Panel) SetLoading() {
	p.todos = nil
	p.list.Clear()
	p.list.AddItem("[gray]Loading todos...", "", 0, nil)
}

func (p *Panel) SetTodos(todos []store.Todo) {
	selectedIndex := p.list.GetCurrentItem()
	selectedID := p.selectedID()

	p.todos = append([]store.Todo(nil), todos...)
	p.list.Clear()

	if len(todos) == 0 {
		p.list.AddItem("[gray]No todos yet. Press a to add one.", "", 0, nil)
		return
	}

	for _, todo := range todos {
		check := "[ ]"
		if todo.Done {
			check = "[x]"
		}

		p.list.AddItem(fmt.Sprintf("%s %s", check, tview.Escape(todo.Title)), "", 0, nil)
	}

	if selectedID != "" && p.SelectTodo(selectedID) {
		return
	}
	if selectedIndex >= len(todos) {
		selectedIndex = len(todos) - 1
	}
	if selectedIndex < 0 {
		selectedIndex = 0
	}
	p.list.SetCurrentItem(selectedIndex)
}

func (p *Panel) SetError(err error) {
	p.todos = nil
	p.list.Clear()
	p.list.AddItem("[red]Could not load todos", err.Error(), 0, nil)
}

func (p *Panel) SelectedTodo() (store.Todo, bool) {
	index := p.list.GetCurrentItem()
	if index < 0 || index >= len(p.todos) {
		return store.Todo{}, false
	}

	return p.todos[index], true
}

func (p *Panel) SelectTodo(id string) bool {
	for index, todo := range p.todos {
		if todo.ID == id {
			p.list.SetCurrentItem(index)
			return true
		}
	}

	return false
}

func (p *Panel) selectedID() string {
	todo, ok := p.SelectedTodo()
	if !ok {
		return ""
	}

	return todo.ID
}
