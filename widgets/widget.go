package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type Widget interface {
	Draw()
	Resize()
	HandleEvent(ev tcell.Event)
	HandleUiEvent(ev AppEvent)
	EmitUiEvent(ev AppEvent)
	SetViewPort(view *views.ViewPort)
	SetParent(w Widget)
	Size() (int, int)
}

type BaseWidget struct {
	parent Widget
}

func (b *BaseWidget) SetParent(w Widget) {
	b.parent = w
}
func (b *BaseWidget) Resize() {
}
func (b *BaseWidget) Size() (int, int) {
	return 0, 0
}
func (b *BaseWidget) EmitUiEvent(ev AppEvent) {
	// XXX check if parent is nil?
	b.parent.HandleUiEvent(ev)
}
