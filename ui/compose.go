package ui

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	sm "github.com/stregouet/nuntius/statesmachines"
	"github.com/stregouet/nuntius/widgets"
	"github.com/stregouet/nuntius/workers"
)

type ComposeView struct {
	machine  *lib.Machine
	bindings config.Mapping
	term     *widgets.Terminal
	screen   tcell.Screen
	*widgets.BaseWidget
}

func NewComposeView(acc *config.Account, bindings config.Mapping, msgCb func(msg string, args ...interface{})) *ComposeView {
	email, err := ioutil.TempFile("", "nuntius-*.eml")
	machine := sm.NewComposeMachine(email)
	if err != nil {
		App.logger.Errorf("cannot create tmp file %v", err)
		machine.Send(&lib.Event{sm.TR_COMPOSE_SET_ERR, nil})
	}
	b := &widgets.BaseWidget{}
	c := &ComposeView{
		machine:    machine,
		bindings:   bindings,
		BaseWidget: b,
	}
	b.OnMessage(msgCb)
	c.setTerminal()
	machine.OnTransition(func(s lib.StateType, ctx interface{}, ev *lib.Event) {
		switch ev.Transition {
		case sm.TR_COMPOSE_SEND:
			state := ctx.(*sm.ComposeMachineCtx)
			App.PostImapMessage(
				&workers.SendMail{
					Body: strings.NewReader(state.Body),
				},
				acc.Name,
				func(response workers.Message) error {
					if r, ok := response.(*workers.Error); ok {
						App.logger.Errorf("cannot send mail %v", r.Error)
						c.Messagef("error sending mail (%v)", r.Error)
					}
					return nil
				},
			)
			c.AskRedraw()
		case sm.TR_COMPOSE_SET_ERR:
			c.AskRedraw()
		case sm.TR_COMPOSE_REVIEW:
			if c.term != nil {
				c.term.Destroy()
				c.term = nil
			}
			c.AskRedraw()
		case sm.TR_COMPOSE_WRITE:
			c.setTerminal()
			c.setTermView(c.GetViewPort(), c.screen)
			c.AskRedraw()
		}
	})
	c.OnSetViewPort(func(view *views.ViewPort, screen tcell.Screen) {
		c.screen = screen
		c.setTermView(view, screen)
	})
	return c
}

func (c *ComposeView) setTermView(view *views.ViewPort, screen tcell.Screen) {
	if c.term != nil {
		c.term.SetViewPort(view, screen)
	}
}

func (c *ComposeView) state() *sm.ComposeMachineCtx {
	return c.machine.Context.(*sm.ComposeMachineCtx)
}

func (c *ComposeView) termClosed(err error) {
	App.transitions <- &lib.Event{sm.TR_COMPOSE_REVIEW, nil}
}

func (c *ComposeView) setTerminal() {
	state := c.state()
	editorName := os.Getenv("EDITOR")
	editor := exec.Command("/bin/sh", "-c", editorName+" "+state.MailFile.Name())
	c.term = widgets.NewTerminal(editor)
	c.term.OnClose = c.termClosed
	c.term.AskingRedraw(func() {
		c.AskRedraw()
	})
}


func (c *ComposeView) err(msg string, format ...interface{}) {
	App.logger.Errorf(msg, format...)
	c.machine.Send(&lib.Event{sm.TR_COMPOSE_SET_ERR, nil})
}

func (c *ComposeView) drawMail(content string) {
	style := tcell.StyleDefault
	bold := style.Bold(true)
	r := strings.NewReader(content)
	msg, err := message.Read(r)
	if err != nil {
		c.Messagef("malformed mail")
		c.err("cannot parse mail %v", err)
		return
	}
	hf := msg.Header.Fields()
	line := 0
	for hf.Next() {
		offset := c.Print(0, line, bold, hf.Key() + ": ")
		val, err := hf.Text()
		if err != nil {
			c.err("cannot parse header field %v", err)
			return
		}
		c.Print(offset, line, style, val)
		line++
	}

	s := bufio.NewScanner(msg.Body)
	line++
	for s.Scan() {
		c.Print(0, line, style, s.Text())
		line++
	}
}

func (c *ComposeView) Draw() {
	style := tcell.StyleDefault
	if c.machine.Current == sm.STATE_COMPOSE_ERR {
		c.Clear()
		c.Print(0, 0, style, "error occured...")
	} else if c.machine.Current == sm.STATE_COMPOSE_WRITE_MAIL {
		c.term.Draw()
	} else {
		c.Clear()
		state := c.state()
		c.drawMail(state.Body)
	}
}

func (c *ComposeView) SetTermView(view *views.ViewPort, screen tcell.Screen) {
	c.term.SetViewPort(view, screen)
}

func (c *ComposeView) IsActiveTerm() bool {
	return c.machine.Current == sm.STATE_COMPOSE_WRITE_MAIL
}

func (c *ComposeView) HandleEvent(ks []*lib.KeyStroke) bool {
	if c.machine.Current == sm.STATE_COMPOSE_WRITE_MAIL {
		return c.term.HandleEvent(ks)
	}
	if cmd := c.bindings.FindCommand(ks); cmd != "" {
		mev, err := c.machine.BuildEvent(cmd)
		if err != nil {
			App.logger.Errorf("error building machine event from `%s` (%v)", cmd, err)
			return false
		}
		if c.machine.Send(mev) {
			return true
		}
	}
	return false
}

func (c *ComposeView) HandleTransitions (ev *lib.Event) bool {
	return c.machine.Send(ev)
}
