package imap

import (
	"github.com/pkg/errors"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/workers"
)

type ImapWorker struct {
	Accounts  map[string]*Account
	requests  chan workers.Message
	responses chan workers.Message

	logger *lib.Logger
}

func NewImapWorker(l *lib.Logger, cfg []*config.Account) *ImapWorker {
	accounts := make(map[string]*Account)
	for _, c := range cfg {
		accounts[c.Name] = NewAccount(l, c)
	}
	return &ImapWorker{
		requests:  make(chan workers.Message, 10),
		responses: make(chan workers.Message, 10),
		Accounts:  accounts,
		logger:    l,
	}
}

func (iw *ImapWorker) Responses() <-chan workers.Message {
	return iw.responses
}
func (iw *ImapWorker) PostMessage(m workers.Message) {
	iw.requests <- m
}

func (iw *ImapWorker) terminate() {
	for _, acc := range iw.Accounts {
		close(acc.requests)
	}
}
func (iw *ImapWorker) postResponse(msg workers.Message, id int) {
	msg.SetId(id)
	iw.responses <- msg
}

func (iw *ImapWorker) Run() {
	defer lib.Recover(iw.logger, nil)
	for _, acc := range iw.Accounts {
		go acc.Run(iw.responses)
	}
	defer iw.terminate()
	for {
		select {
		case msg := <-iw.requests:
			if acc, ok := iw.Accounts[msg.GetAccName()]; ok {
				acc.requests <- msg
			} else {
				r := &workers.Error{
					Error: errors.Errorf("no worker found for imap account `%s`", msg.GetAccName()),
				}
				iw.postResponse(r, msg.GetId())
			}
		}
	}
}
