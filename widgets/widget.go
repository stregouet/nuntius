package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/mattn/go-runewidth"

	"github.com/stregouet/nuntius/lib"
)

type Widget interface {
	HandleEvent(ks []*lib.KeyStroke) bool
	Draw()
	SetViewPort(view *views.ViewPort, screen tcell.Screen)
	GetViewPort() *views.ViewPort
	SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style)
	IsActiveTerm() bool
	HandleTransitions(ev *lib.Event) bool

	AskRedraw()
	AskingRedraw(func())
	OnMessage(f func(msg string, args ...interface{}))

	Resize()
	Size() (int, int)
}

type BaseWidget struct {
	redrawCb func()
	viewCb   func(view *views.ViewPort, screen tcell.Screen)
	msgCb    func(msg string, args ...interface{})
	view     *views.ViewPort
	screen   tcell.Screen
}

func (b *BaseWidget) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) {
	b.view.SetContent(x, y, mainc, combc, style)
}
func (b *BaseWidget) ScrollUp(rows int) {
	b.view.ScrollUp(rows)
}
func (b *BaseWidget) ScrollDown(rows int) {
	b.view.ScrollDown(rows)
}
func (b *BaseWidget) SetViewPort(view *views.ViewPort, screen tcell.Screen) {
	b.screen = screen
	b.view = view
	if b.viewCb != nil {
		b.viewCb(view, screen)
	}
}
func (b *BaseWidget) GetViewPort() *views.ViewPort {
	return b.view
}

func (b *BaseWidget) Clear() {
	if b.view != nil {
		b.view.Clear()
	}
}

func (b *BaseWidget) OnSetViewPort(f func(view *views.ViewPort, screen tcell.Screen)) {
	b.viewCb = f
}

func (b *BaseWidget) OnMessage(f func(msg string, args ...interface{})) {
	b.msgCb = f
}

func (b *BaseWidget) Messagef(msg string, args ...interface{}) {
	if b.msgCb != nil {
		b.msgCb(msg, args...)
	}
}

func (b *BaseWidget) AskingRedraw(f func()) {
	b.redrawCb = f
}
func (b *BaseWidget) AskRedraw() {
	if b.redrawCb != nil {
		b.redrawCb()
	}
}
func (b *BaseWidget) Resize() {
}
func (b *BaseWidget) Size() (int, int) {
	return 0, 0
}

func (b *BaseWidget) HideCursor() {
	if b.screen != nil {
		b.screen.HideCursor()
	}
}
func (b *BaseWidget) ShowCursor(x int, y int) {
	if b.screen != nil && b.view != nil {
		physx, physy, _, _ := b.view.GetPhysical()
		b.screen.ShowCursor(x+physx, y+physy)
	}
}
func (b *BaseWidget) IsActiveTerm() bool {
	return false
}

func (b *BaseWidget) HandleTransitions(ev *lib.Event) bool {
	return false
}

func (b *BaseWidget) Print(x, y int, style tcell.Style, str string) int {
	width, height := b.view.Size()

	old_x := x

	newline := func() bool {
		x = old_x
		y++
		return y < height
	}
	for _, ch := range str {
		switch ch {
		case '\n':
			if !newline() {
				return runewidth.StringWidth(str)
			}
		case '\r':
			x = old_x
		default:
			crunes := []rune{}
			b.SetContent(x, y, ch, crunes, style)
			x += runewidth.RuneWidth(ch)
			if x == old_x+width {
				if !newline() {
					return runewidth.StringWidth(str)
				}
			}
		}
	}

	return runewidth.StringWidth(str)
}
