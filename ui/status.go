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
		Text:       &widgets.Text{},
		tmpContent: lib.NewConcurrentList(make([]interface{}, 0)),
	}
	s.SetContent(msg)
	return s
}

func (s *Status) showMessage(msg string, style tcell.Style) {
	c := &ContentWithStyle{
		content: msg,
		style:   style,
	}
	s.tmpContent.Push(c)
	s.AskRedraw()
	go func() {
		time.Sleep(5 * time.Second)
		s.tmpContent.Remove(c)
		s.AskRedraw()
	}()
}

func (s *Status) ShowMessage(msg string) {
	s.showMessage(msg, tcell.StyleDefault)
}

func (s *Status) ShowError(msg string) {
	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed)
	s.showMessage(msg, style)
}

func (s *Status) Draw() {
	s.Clear()
	style := tcell.StyleDefault
	content := s.GetContent()
	if s.tmpContent.Length() > 0 {
		withstyle := s.tmpContent.Last().(*ContentWithStyle)
		content = withstyle.content
		style = withstyle.style
	}
	s.Print(0, 0, style, content)
}
