package widgets

import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
)

const ARROW = '➤'

type ITreeLine interface {
	IRune
	Depth() int
}

type TreeWidget struct {
	lines      []ITreeLine
	selected   int
	indentSize int
	OnSelect   func(ITreeLine)
	ListWidget
}

func NewTree() *TreeWidget {
	return &TreeWidget{
		indentSize: 2,
		lines:      make([]ITreeLine, 0),
		OnSelect:   func(ITreeLine) {},
		selected:   1,
	}
}

func NewTreeWithInitSelected(init int) *TreeWidget {
	t := NewTree()
	t.selected = init
	return t
}

func (t *TreeWidget) ClearLines() {
	t.lines = make([]ITreeLine, 0)
}

func (t *TreeWidget) AddLine(line ITreeLine) {
	t.lines = append(t.lines, line)
}

func arrayWithSpace(length int) []rune {
	result := make([]rune, length)
	for i, _ := range result {
		result[i] = ' '
	}
	return result
}

func samelevelInNextlines(nextlines []ITreeLine, level int) bool {
	for _, nextline := range nextlines {
		if nextline.Depth() == level {
			return true
		} else if nextline.Depth() < level {
			return false
		}
	}
	return false
}

func (t *TreeWidget) Draw() {
	t.Clear()
	v := t.view
	w, h := v.Size()
	for y, line := range t.lines {
		linenum := y + 1
		if linenum > h {
			break
		}

		arrowCells := arrayWithSpace(line.Depth() * t.indentSize)
		if y != 0 && len(arrowCells) > 0 {
			nextlines := t.lines[y+1:]
			for level := 1; level < line.Depth(); level++ {
				if samelevelInNextlines(nextlines, level) {
					arrowCells[(level-1)*t.indentSize] = tcell.RuneVLine
				}
			}
			angleIdx := (line.Depth() - 1) * t.indentSize
			if samelevelInNextlines(nextlines, line.Depth()) {
				arrowCells[angleIdx] = tcell.RuneLTee
			} else {
				arrowCells[angleIdx] = tcell.RuneLLCorner
			}
			for i := angleIdx + 1; i < len(arrowCells); i++ {
				if i == (len(arrowCells) - 1) {
					arrowCells[i] = ARROW
				} else {
					arrowCells[i] = tcell.RuneHLine
				}
			}
		}

		style := tcell.StyleDefault
		for x, c := range arrowCells {
			colnum := x + 1
			if colnum > w {
				break
			}
			v.SetContent(x, y, c, nil, style)
		}
		if linenum == t.selected {
			style = style.Reverse(true)
		}
		for x, r := range line.ToRune() {
			x = len(arrowCells) + x
			colnum := x + 1
			if colnum > w {
				break
			}
			v.SetContent(x, y, r, nil, style)
		}
	}
}

func (t *TreeWidget) onSelect() {
	t.OnSelect(t.GetSelected())
}

func (t *TreeWidget) GetSelected() ITreeLine {
	return t.lines[t.selected-1]
}
func (t *TreeWidget) SetSelected(s int) {
	t.selected = s
	t.AskRedraw()
}

func (t *TreeWidget) HandleEvent(ks []*lib.KeyStroke) bool {
	switch ks[0].Key {
	case tcell.KeyUp, tcell.KeyCtrlP:
		t.SetSelected(max(t.selected-1, 1))
		return true
	case tcell.KeyDown, tcell.KeyCtrlN:
		t.SetSelected(min(t.selected+1, len(t.lines)))
		return true
	case tcell.KeyEnter:
		t.onSelect()
		return true
	}
	return false
}
