package ui

import (
	"github.com/gdamore/tcell/v2"

    "github.com/stregouet/nuntius/models"
    "github.com/stregouet/nuntius/widgets"
)


type MessageList struct {
    *widgets.ListWidget
}

func NewMessageList() *MessageList {
    l := widgets.NewList()
    return &MessageList{
        ListWidget: l,
    }
}

func (ml *MessageList) SetList(newlist []string) {
    ml.ClearLines()
    App.logger.Debugf("setting list with %v", newlist)
    for _, item := range newlist {
        m := &models.Message{Subject: item}
        ml.AddLine(m)
    }
    ml.AskRedraw()
}

func (ml *MessageList) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'Q', 'q':
                App.Stop()
				return
			}
		}
	}
    ml.ListWidget.HandleEvent(ev)
}
