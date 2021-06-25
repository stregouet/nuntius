package ui

import (
	"bufio"
	"errors"
	"io"
	"os"

	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
)

var ErrStopWalk = errors.New("stop walk")

type MailView struct {
	machine  *lib.Machine
	bindings config.Mapping
	partsView *MailPartsView
	*widgets.BaseWidget
}

func NewMailView(bindings config.Mapping, partsBindings config.Mapping, mail *models.Mail) *MailView {
	b := widgets.BaseWidget{}
	machine := sm.NewMailMachine(mail)
	machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		switch ev.Transition {
		case sm.TR_SCROLL_UP_MAIL:
			b.ScrollUp(1)
			b.AskRedraw()
		case sm.TR_SCROLL_DOWN_MAIL:
			b.ScrollDown(1)
			b.AskRedraw()
		case sm.TR_SHOW_MAIL_PARTS, sm.TR_SHOW_MAIL_PART:
			b.AskRedraw()
		}
	})
	mv := &MailView{
		machine:    machine,
		bindings:   bindings,
		BaseWidget: &b,
	}
	mv.partsView = NewMailPartsView(partsBindings, mail.Parts, mv.onSelectPart)
	mv.partsView.AskingRedraw(func() {
		mv.AskRedraw()
	})
	return mv
}

func (mv *MailView) SetPartsView(view *views.ViewPort) {
	mv.partsView.SetViewPort(view)
}

func (mv *MailView) onSelectPart(path models.BodyPath) {
	ev := &lib.Event{sm.TR_SHOW_MAIL_PART, path}
	mv.machine.Send(ev)
}

func (mv *MailView) SetFilepath(filepath string) {
	ev := &lib.Event{sm.TR_SET_FILEPATH, filepath}
	mv.machine.Send(ev)
	mv.AskRedraw()
}

func (mv *MailView) drawHeader(header message.Header, offset int) int {
	style := tcell.StyleDefault
	bold := style.Bold(true)
	line := offset
	for _, key := range []string{"from", "to", "cc", "message-id", "in-reply-to", "subject"} {
		value, err := header.Text(key)
		if value != "" && err == nil {
			col := 0
			for _, c := range []rune(key) {
				mv.SetContent(col, line, c, nil, bold)
				col++
			}
			col += 2
			for _, c := range []rune(value) {
				mv.SetContent(col, line, c, nil, style)
				col++
			}
			line++
		}
	}
	return line
}

func (mv *MailView) drawBody(body io.Reader, lineoffset int) {
	style := tcell.StyleDefault
	s := bufio.NewScanner(body)
	line := lineoffset + 1
	for s.Scan() {
		for col, c := range []rune(s.Text()) {
			mv.SetContent(col, line, c, nil, style)
		}
		line++
	}
}

func (mv *MailView) state() *sm.MailMachineCtx {
	return mv.machine.Context.(*sm.MailMachineCtx)
}

func (mv *MailView) Draw() {
	style := tcell.StyleDefault
	if mv.machine.Current == sm.STATE_LOAD_MAIL {
		for i, c := range "loading..." {
			mv.SetContent(i, 0, c, nil, style)
		}
	} else if mv.machine.Current == sm.STATE_SHOW_MAIL_PARTS {
		mv.partsView.Draw()
	} else {
		state := mv.state()
		f, err := os.Open(state.Filepath)
		if err != nil {
			App.logger.Errorf("cannot open filepath %v (filepath: %s)", err, state.Filepath)
			return
		}
		defer f.Close()
		msg, err := message.Read(f)
		if err != nil {
			App.logger.Errorf("cannot build go-message from file %v", err)
			return
		}
		var body io.Reader
		selectedpath, err := state.SelectedPart.ToMessagePath()
		if err != nil {
			App.logger.Errorf("cannot build message path %v", err)
			return
		}
		err = msg.Walk(func(path []int, e *message.Entity, err error) error {
			if lib.IsSliceIntEqual(path, selectedpath) {
				body = e.Body
				return ErrStopWalk
			}
			return err
		})
		if err != nil && err != ErrStopWalk {
			App.logger.Errorf("cannot walk in message parts %v", err)
			return
		}
		if body != nil {
			offset := mv.drawHeader(msg.Header, 0)
			mv.drawBody(body, offset)
		} else {
			App.logger.Debug("cannot find text/plain body")
			for i, c := range state.Filepath {
				mv.SetContent(i, 0, c, nil, style)
			}
		}
	}
}

func (mv *MailView) HandleEvent(ks []*lib.KeyStroke) bool {
	if mv.machine.Current == sm.STATE_SHOW_MAIL_PARTS {
		return mv.partsView.HandleEvent(ks)
	}
	if cmd := mv.bindings.FindCommand(ks); cmd != "" {
		mev, err := mv.machine.BuildEvent(cmd)
		if err != nil {
			App.logger.Errorf("error building machine event from `%s` (%v)", cmd, err)
			return false
		}
		if mv.machine.Send(mev) {
			return true
		}
	}
	return false
}
