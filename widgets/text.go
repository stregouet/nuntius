package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type Text struct {
	content string
	view    *views.ViewPort
	BaseWidget
}

func (t *Text) SetContent(content string) {
	t.content = content
}

func (t *Text) Draw() {
	style := tcell.StyleDefault
	for i, ch := range t.content {
		t.view.SetContent(i, 0, ch, nil, style)
	}
}
func (t *Text) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRight:
			t.AskRedraw()
			return
		case tcell.KeyLeft:
			t.view.ScrollLeft(1)
			t.AskRedraw()
			return
		}
	}
}
func (t *Text) SetViewPort(view *views.ViewPort) {
	t.view = view
}
