package widgets

import (
	"github.com/gdamore/tcell/v2"
)

type Text struct {
	content string
	BaseWidget
}

func (t *Text) SetContent(content string) {
	t.content = content
}

func (t *Text) GetContent() string {
	return t.content
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
