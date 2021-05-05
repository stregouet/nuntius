package widgets

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type IRune interface {
	ToRune() []rune
}

type ListWidget struct {
	lines    []IRune
	selected int
	OnSelect func(IRune)
	view     views.View
	logger   *log.Logger
	views.WidgetWatchers
}

func NewList(l *log.Logger) ListWidget {
	return ListWidget{
		lines:    make([]IRune, 0),
		OnSelect: func(IRune) {},
		selected: 1,
		logger:   l,
	}
}

func (l *ListWidget) AddLine(line IRune) {
	l.lines = append(l.lines, line)
}

func (l *ListWidget) SetView(view views.View) {
	l.view = view
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
func (l *ListWidget) Resize() {}
func (l *ListWidget) Size() (int, int) {
	w, h := 0, 0
	for _, line := range l.lines {
		lw := len(line.ToRune())
		if lw > w {
			w = lw
		}
		h += 1
	}
	return w, h
}

func (l *ListWidget) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyUp, tcell.KeyCtrlP:
			l.selected = max(l.selected-1, 1)
			return true
		case tcell.KeyDown, tcell.KeyCtrlN:
			l.selected = min(l.selected+1, len(l.lines))
			return true
		case tcell.KeyEnter:
			l.onSelect()
			return true
		}
	}
	return false
}

func (l *ListWidget) onSelect() {
	l.OnSelect(l.lines[l.selected-1])
}
