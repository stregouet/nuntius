package ui

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/widgets"
)

type Window struct {
	logger  *log.Logger
	screen  tcell.Screen
	content widgets.Widget
	app     *Application
}


func NewWindow(app *Application, l *log.Logger) *Window {
	return &Window{app: app, logger: l}
}

func (w *Window) SetContent(widget widgets.Widget) {
	w.content = widget
	widget.SetParent(w)
	if w.screen != nil {
		vport := views.NewViewPort(w.screen, 0, 0, -1, -1)
		w.content.SetViewPort(vport)
	}
}
func (w *Window) SetParent(_ widgets.Widget)            {}
func (w *Window) SetViewPort(v *views.ViewPort) {}
func (w *Window) EmitUiEvent(ev widgets.AppEvent) {
	// XXX check if parent is nil?
	w.app.HandleUiEvent(ev)
}
func (w *Window) HandleUiEvent(ev widgets.AppEvent) {
	switch ev {
	case widgets.REDRAW_EVENT:
		w.Redraw()
	default:
		w.EmitUiEvent(ev)
	}
}

func (w *Window) Redraw() {
	w.Draw()
	w.screen.Show()
}

func (w *Window) Resize()          {}
func (w *Window) Size() (int, int) { return w.screen.Size() }
func (w *Window) SetScreen(s tcell.Screen) {
	w.screen = s
	vport := views.NewViewPort(s, 0, 0, -1, -1)
	w.content.SetViewPort(vport)
}
func (w *Window) Draw() {
	w.screen.Clear()
	w.content.Draw()
}

func (w *Window) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
				w.EmitUiEvent(widgets.QUIT_EVENT)
				return
			}
		}
	}
	w.content.HandleEvent(ev)
}
