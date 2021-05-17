package widgets

import (
	"github.com/gdamore/tcell/v2"
)

type IRune interface {
	ToRune() []rune
}

type ListWidget struct {
	lines    []IRune
	selected int
	OnSelect func(IRune)
	BaseWidget
}

func NewList() *ListWidget {
	return &ListWidget{
		lines:    make([]IRune, 0),
		OnSelect: func(IRune) {},
		selected: 1,
	}
}

func (l *ListWidget) AddLine(line IRune) {
	l.lines = append(l.lines, line)
}

func (l *ListWidget) ClearLines() {
	l.lines = make([]IRune, 0)
}


func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
func (l *ListWidget) Draw() {
	v := l.view
	w, h := v.Size()
	for y, line := range l.lines {
		linenum := y + 1
		if linenum > h {
			break
		}
		style := tcell.StyleDefault
		if linenum == l.selected {
			style = style.Reverse(true)
		}
		for x, r := range line.ToRune() {
			colnum := x + 1
			if colnum > w {
				break
			}
			v.SetContent(x, y, r, nil, style)
		}
	}
}

func (l *ListWidget) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyUp, tcell.KeyCtrlP:
			l.selected = max(l.selected-1, 1)
			l.AskRedraw()
			return
		case tcell.KeyDown, tcell.KeyCtrlN:
			l.selected = min(l.selected+1, len(l.lines))
			l.AskRedraw()
			return
		case tcell.KeyEnter:
			l.onSelect()
			return
		}
	}
}

func (l *ListWidget) onSelect() {
	l.OnSelect(l.lines[l.selected-1])
}
