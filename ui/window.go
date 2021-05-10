package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/widgets"
)

type Window struct {
	screen tcell.Screen

	selectedTab int
	tabs        []widgets.Widget
	ex          *widgets.Text
}

func NewWindow() *Window {
	t := &widgets.Text{}
	t.SetContent("ici c'est pour les commandes")
	w := &Window{
		tabs: make([]widgets.Widget, 0),
		ex:   t,
	}
	m := NewMailboxesView([]string{"inbox", "junk"})
	w.AddTab(m)
	return w
}

func (w *Window) tabViewPort() *views.ViewPort {
	_, h := w.screen.Size()
	return views.NewViewPort(w.screen, 0, 0, -1, h-1)
}
func (w *Window) exViewPort() *views.ViewPort {
	_, h := w.screen.Size()
	return views.NewViewPort(w.screen, 0, h-1, -1, -1)
}

func (w *Window) AddTab(widget widgets.Widget) {
	w.tabs = append(w.tabs, widget)
	if w.screen != nil {
		widget.SetViewPort(w.tabViewPort())
	}
}

func (w *Window) Redraw() {
	w.Draw()
	w.screen.Show()
}

func (w *Window) Size() (int, int) { return w.screen.Size() }
func (w *Window) SetScreen(s tcell.Screen) {
	w.screen = s
	w.ex.SetViewPort(w.exViewPort())
	for _, t := range w.tabs {
		t.SetViewPort(w.tabViewPort())
	}
}
func (w *Window) Draw() {
	w.tabs[w.selectedTab].Draw()
	w.ex.Draw()
}
func (w *Window) TabHandleEvent(ev tcell.Event) {
	w.tabs[w.selectedTab].HandleEvent(ev)
	if w.tabs[w.selectedTab].ShouldRedraw() {
		w.Redraw()
	}
}
func (w *Window) ExHandleEvent(ev tcell.Event) {
	w.ex.HandleEvent(ev)
	if w.ex.ShouldRedraw() {
		w.Redraw()
	}
}
