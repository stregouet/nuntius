package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type Widget interface {
	HandleEvent(ev tcell.Event)
	Draw()
	SetViewPort(view *views.ViewPort)
	GetViewPort() *views.ViewPort

	AskRedraw()
	AskingRedraw(func())

	Resize()
	Size() (int, int)
}

type BaseWidget struct {
	redrawCb func()
	view    *views.ViewPort
}

func (b *BaseWidget) SetViewPort(view *views.ViewPort) {
	b.view = view
}
func (b *BaseWidget) GetViewPort() *views.ViewPort {
	return b.view
}

func (b *BaseWidget) Clear() {
	b.view.Clear()
}

func (b *BaseWidget) AskingRedraw(f func()) {
	b.redrawCb = f
}
func (b* BaseWidget) AskRedraw() {
	b.Clear()
	if b.redrawCb != nil {
		b.redrawCb()
	}
}
func (b *BaseWidget) Resize() {
}
func (b *BaseWidget) Size() (int, int) {
	return 0, 0
}
