package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/widgets"
)

type ContentWithStyle struct {
	content string
	style   tcell.Style
}

type Status struct {
	tmpContent *lib.ConcurrentList
	*widgets.Text
}

func NewStatus(msg string) *Status {
	s := &Status{
		Text: &widgets.Text{},
		tmpContent: lib.NewConcurrentList(make([]interface{}, 0)),
	}
	s.SetContent(msg)
	return s
}

func (s *Status) ShowMessage(msg string) {
	c := &ContentWithStyle{
		content: msg,
		style:   tcell.StyleDefault,
	}
	s.tmpContent.Push(c)
	s.AskRedraw()
	go func() {
		time.Sleep(5 * time.Second)
		s.tmpContent.Remove(c)
		s.AskRedraw()
	}()
}

func (s *Status) Draw() {
	view := s.GetViewPort()
	style := tcell.StyleDefault
	content := s.GetContent()
	if s.tmpContent.Length() > 0 {
		withstyle := s.tmpContent.Last().(*ContentWithStyle)
		content = withstyle.content
		style = withstyle.style
	}
	for i, ch := range content {
		view.SetContent(i, 0, ch, nil, style)
	}
}
