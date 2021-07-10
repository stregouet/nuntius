package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/lib"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
)

type Status struct {
	tmpContent *lib.ConcurrentList
	machine    *lib.Machine
	onEndCmdCb func(string)
	*widgets.Text
}

func NewStatus(msg string, onEndCmd func(string)) *Status {
	machine := sm.NewStatusMachine()
	s := &Status{
		onEndCmdCb: onEndCmd,
		machine:    machine,
		Text:       &widgets.Text{},
		tmpContent: lib.NewConcurrentList(make([]interface{}, 0)),
	}
	s.machine.OnTransition(func(state lib.StateType, ctx interface{}, ev *lib.Event) {
		switch ev.Transition {
		case sm.TR_STATUS_START_WRITING, sm.TR_STATUS_WRITE_CHAR, sm.TR_STATUS_MOVE_CURSOR, sm.TR_STATUS_RM_CHAR, sm.TR_STATUS_RM_WORD, sm.TR_STATUS_BROWSE_HISTORY:
			s.AskRedraw()
		case sm.TR_STATUS_VALIDATE:
			c := ctx.(*sm.StatusMachineCtx)
			s.onEndCmdCb(c.History[len(c.History)-1])
			s.AskRedraw()
		case sm.TR_STATUS_CANCEL:
			s.onEndCmdCb("")
			s.AskRedraw()
		}
	})
	s.SetContent(msg)
	return s
}

func (s *Status) showMessage(msg string, style tcell.Style) {
	c := &widgets.ContentWithStyle{
		Content: msg,
		Style:   style,
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

func (s *Status) state() *sm.StatusMachineCtx {
	return s.machine.Context.(*sm.StatusMachineCtx)
}

func (s *Status) Draw() {
	s.Clear()
	style := tcell.StyleDefault
	if s.machine.Current == sm.STATE_STATUS_WRITE_CMD {
		state := s.state()
		s.ShowCursor(state.CursorPos+1, 0)
		s.Print(0, 0, style, ":"+string(state.WriteContent))
	} else {
		content := s.GetContent()
		if s.tmpContent.Length() > 0 {
			withstyle := s.tmpContent.Last().(*widgets.ContentWithStyle)
			content = withstyle.Content
			style = withstyle.Style
		}
		s.Print(0, 0, style, content)
	}
}

func (s *Status) HandleEvent(ks []*lib.KeyStroke) bool {
	if ks[0].Rune != 0 {
		s.machine.Send(&lib.Event{sm.TR_STATUS_WRITE_CHAR, ks[0].Rune})
		return true
	}
	switch ks[0].Key {
	case tcell.KeyESC:
		s.machine.Send(&lib.Event{sm.TR_STATUS_CANCEL, nil})
		return true
	case tcell.KeyLeft:
		s.machine.Send(&lib.Event{sm.TR_STATUS_MOVE_CURSOR, -1})
		return true
	case tcell.KeyRight:
		s.machine.Send(&lib.Event{sm.TR_STATUS_MOVE_CURSOR, 1})
		return true
	case tcell.KeyUp:
		s.machine.Send(&lib.Event{sm.TR_STATUS_BROWSE_HISTORY, -1})
		return true
	case tcell.KeyDown:
		s.machine.Send(&lib.Event{sm.TR_STATUS_BROWSE_HISTORY, 1})
		return true
	case tcell.KeyEnter:
		s.machine.Send(&lib.Event{sm.TR_STATUS_VALIDATE, nil})
		return true
	case tcell.KeyBackspace2:
		var mev lib.Event
		evk := ks[0].Tev.(*tcell.EventKey)
		if evk.Modifiers() == tcell.ModAlt {
			mev = lib.Event{sm.TR_STATUS_RM_WORD, nil}
		} else {
			mev = lib.Event{sm.TR_STATUS_RM_CHAR, nil}
		}
		s.machine.Send(&mev)
		return true
	}
	return false
}
