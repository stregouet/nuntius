package ui

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/gdamore/tcell/v2"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
	"github.com/stregouet/nuntius/workers"
)

var ErrStopWalk = errors.New("stop walk")

type MailView struct {
	machine   *lib.Machine
	bindings  config.Mapping
	partsView *MailPartsView
	filters   config.Filters
	*widgets.BaseWidget
}

func NewMailView(bindings config.Mapping, partsBindings config.Mapping, filters config.Filters) *MailView {
	b := widgets.BaseWidget{}
	machine := sm.NewMailMachine()
	mv := &MailView{
		machine:    machine,
		bindings:   bindings,
		BaseWidget: &b,
		filters:    filters,
	}
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
		case sm.TR_SET_MAIL:
			state := ctx.(*sm.MailMachineCtx)
			mv.partsView = NewMailPartsView(partsBindings, state.Mail.Parts, mv.onSelectPart)
			mv.partsView.AskingRedraw(func() {
				mv.AskRedraw()
			})
		}
	})
	return mv
}

// Tab interface
func (mv *MailView) TabTitle() string {
	var sub string
	if mv.machine.Current == sm.STATE_LOAD_MAIL {
		sub = "load mail…"
	} else {
		s := mv.state()
		sub = s.Mail.Subject
	}
	return "\uf0e0 " + sub
}

func (mv *MailView) SetMail(m *models.Mail, mailbox, acc string) {
	ev := &lib.Event{sm.TR_SET_MAIL, m}
	mv.machine.Send(ev)
	App.PostImapMessage(
		&workers.FetchFullMail{Uid: m.Uid, Mailbox: mailbox},
		acc,
		func(response workers.Message) error {
			switch r := response.(type) {
			case *workers.Error:
				mv.Messagef("error fetching mail content %v", r.Error)
			case *workers.FetchFullMailRes:
				mv.SetFilepath(r.Filepath)
				App.logger.Debugf("full mail received, `%v`", r.Filepath)
			default:
				App.logger.Error("unknown response type")
			}
			return nil
		})
}

func (mv *MailView) onSelectPart(part *models.BodyPart) {
	ev := &lib.Event{sm.TR_SHOW_MAIL_PART, part}
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
			col := mv.Print(0, line, bold, key)
			col += 2
			mv.Print(col, line, style, value)
			line++
		}
	}
	return line
}

func (mv *MailView) drawBody(mailbody io.Reader, lineoffset int) {
	style := tcell.StyleDefault
	line := lineoffset + 1
	state := mv.state()
	var body io.Reader
	filter := state.SelectedPart.FindMatch(mv.filters)
	if filter != "" {
		cmd := exec.Command("sh", "-c", filter)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			App.logger.Errorf("error running cmd %v", err)
			mv.Print(0, line, style, "error running cmd")
			return
		}
		go func() {
			defer stdin.Close()
			io.Copy(stdin, mailbody)
		}()
		out, err := cmd.CombinedOutput()
		if err != nil {
			App.logger.Errorf("error running cmd %v %s", err, out)
			mv.Print(0, line, style, "error running cmd")
			return
		}
		body = bytes.NewBuffer(out)
	} else {
		body = mailbody
	}
	s := bufio.NewScanner(body)
	for s.Scan() {
		mv.Print(0, line, style, s.Text())
		line++
	}
}

func (mv *MailView) state() *sm.MailMachineCtx {
	return mv.machine.Context.(*sm.MailMachineCtx)
}

func (mv *MailView) Draw() {
	mv.Clear()
	style := tcell.StyleDefault
	if mv.machine.Current == sm.STATE_LOAD_MAIL {
		mv.Print(0, 0, style, "loading...")
	} else if mv.machine.Current == sm.STATE_SHOW_MAIL_PARTS {
		if mv.partsView.GetViewPort() == nil {
			mv.partsView.SetViewPort(mv.GetViewPort(), nil)
		}
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
		selectedpath, err := state.SelectedPart.Path.ToMessagePath()
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
			App.logger.Debugf("cannot find selected part %v", state.SelectedPart)
			mv.Print(0, 0, style, "no body for selected part (see mail at: "+state.Filepath+")")
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
