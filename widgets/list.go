package widgets

import (
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
)

type ContentWithStyle struct {
	Content string
	Style   tcell.Style
}

func NewContent(c string) *ContentWithStyle {
	return &ContentWithStyle{c, tcell.StyleDefault}
}

func (cs *ContentWithStyle) Reverse(should bool) tcell.Style {
	return cs.Style.Reverse(should)
}

type IStyled interface {
	StyledContent() []*ContentWithStyle
}

type ListWidget struct {
	lines             []IStyled
	selected          int
	OnSelect          func(IStyled)
	viewableFirstLine int
	BaseWidget
}

func NewList() *ListWidget {
	return &ListWidget{
		lines:    make([]IStyled, 0),
		OnSelect: func(IStyled) {},
		selected: 1,
	}
}

func (l *ListWidget) AddLine(line IStyled) {
	l.lines = append(l.lines, line)
}

func (l *ListWidget) ClearLines() {
	l.lines = make([]IStyled, 0)
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

func (l *ListWidget) SetSelected(s int) {
	l.selected = s
	_, h := l.view.Size()
	if s > (h + l.viewableFirstLine) {
		diff := s - (h + l.viewableFirstLine)
		l.ScrollDown(diff)
		l.viewableFirstLine += diff
	} else if s < (l.viewableFirstLine + 1) {
		diff := (l.viewableFirstLine + 1) - s
		l.ScrollUp(diff)
		l.viewableFirstLine -= diff
	}
	l.AskRedraw()
}

func (l *ListWidget) Draw() {
	v := l.view
	_, h := v.Size()
	for y, line := range l.lines {
		linenum := y + 1
		if linenum > (h + 1000000) {
			break
		}
		coloffset := 0
		for _, withstyle := range line.StyledContent() {
			coloffset += l.Print(
				coloffset,
				y,
				withstyle.Reverse(linenum == l.selected),
				withstyle.Content,
			)
		}
	}
}

func (l *ListWidget) HandleEvent(ks []*lib.KeyStroke) bool {
	switch ks[0].Key {
	case tcell.KeyUp, tcell.KeyCtrlP:
		l.selected = max(l.selected-1, 1)
		l.AskRedraw()
		return true
	case tcell.KeyDown, tcell.KeyCtrlN:
		l.selected = min(l.selected+1, len(l.lines))
		l.AskRedraw()
		return true
	case tcell.KeyEnter:
		l.onSelect()
		return true
	}
	return false
}

func (l *ListWidget) onSelect() {
	l.OnSelect(l.lines[l.selected-1])
}
