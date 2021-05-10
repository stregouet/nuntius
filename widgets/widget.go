package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type Widget interface {
	HandleEvent(ev tcell.Event)
	Draw()
	SetViewPort(view *views.ViewPort)

	ShouldRedraw() bool
	AskRedraw()
	ResetRedraw()

	Resize()
	Size() (int, int)
}

type BaseWidget struct {
	triggerRedraw bool
}


func (b *BaseWidget) ShouldRedraw() bool {
	return b.triggerRedraw
}
func (b* BaseWidget) AskRedraw() {
	b.triggerRedraw = true
}
func (b* BaseWidget) ResetRedraw() {
	b.triggerRedraw = false
}
func (b *BaseWidget) Resize() {
}
func (b *BaseWidget) Size() (int, int) {
	return 0, 0
}
