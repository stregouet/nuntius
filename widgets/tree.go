package widgets

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

const ARROW = '➤'
const ANGLE1 = '└'
const ANGLE2 = '├'
const PIPE = '│'
const LINE = '─'

type ITreeLine interface {
	IRune
	Depth() int
}

type TreeWidget struct {
	lines      []ITreeLine
	selected   int
	indentSize int
	OnSelect   func(ITreeLine)
	logger     *log.Logger
	ListWidget
}

func NewTree(l *log.Logger) TreeWidget {
	return TreeWidget{
		indentSize: 2,
		lines:      make([]ITreeLine, 0),
		OnSelect:   func(ITreeLine) {},
		selected:   1,
		logger:     l,
	}
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
	v := t.view
	w, h := v.Size()
	for y, line := range t.lines {
		linenum := y + 1
		if linenum > h {
			break
		}

		arrowCells := arrayWithSpace(line.Depth() * t.indentSize)
		if y != 0 {
			nextlines := t.lines[y+1:]
			for level := 1; level < line.Depth(); level++ {
				if samelevelInNextlines(nextlines, level) {
					arrowCells[(level-1)*t.indentSize] = PIPE
				}
			}
			angleIdx := (line.Depth() - 1) * t.indentSize
			if samelevelInNextlines(nextlines, line.Depth()) {
				arrowCells[angleIdx] = ANGLE2
			} else {
				arrowCells[angleIdx] = ANGLE1
			}
			for i := angleIdx + 1; i < len(arrowCells); i++ {
				if i == (len(arrowCells) - 1) {
					arrowCells[i] = ARROW
				} else {
					arrowCells[i] = LINE
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

func (t *TreeWidget) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyUp, tcell.KeyCtrlP:
			t.selected = max(t.selected-1, 1)
			t.EmitUiEvent(REDRAW_EVENT)
			return
		case tcell.KeyDown, tcell.KeyCtrlN:
			t.selected = min(t.selected+1, len(t.lines))
			t.EmitUiEvent(REDRAW_EVENT)
			return
		case tcell.KeyEnter:
			t.onSelect()
			return
		}
	}
}
