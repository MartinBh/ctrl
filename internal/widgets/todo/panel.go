package todo

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/store"
	"github.com/martinbhatta/ctrl/internal/theme"
)

type Panel struct {
	list *tview.List
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

func (p *Panel) SetTodos(todos []store.Todo) {
	p.list.Clear()

	if len(todos) == 0 {
		p.list.AddItem("[gray]No todos yet.", "", 0, nil)
		return
	}

	for _, todo := range todos {
		check := "[ ]"
		if todo.Done {
			check = "[x]"
		}

		p.list.AddItem(fmt.Sprintf("%s %s", check, todo.Title), "", 0, nil)
	}
}

func (p *Panel) SetError(err error) {
	p.list.Clear()
	p.list.AddItem("[red]Could not load todos", err.Error(), 0, nil)
}
