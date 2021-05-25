package imap

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/pkg/errors"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/workers"
)

type Account struct {
	cfg      *config.Account
	requests chan workers.Message
	c        *client.Client
	logger   *lib.Logger
}

func NewAccount(l *lib.Logger, c *config.Account) *Account {
	return &Account{
		cfg:      c,
		requests: make(chan workers.Message, 10),
		logger:   l,
	}
}

func (a *Account) getPass() (string, error) {
	out, err := exec.Command("sh", "-c", a.cfg.Imap.PassCmd).Output()
	if err != nil {
		return "", errors.Wrap(err, "cannot exec passcmd")
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (a *Account) connect() error {
	var err error
	if a.cfg.Imap.Tls {
		a.c, err = client.DialTLS(fmt.Sprintf("%s:%d", a.cfg.Imap.Host, a.cfg.Imap.Port), nil)
	}
	p, err := a.getPass()
	if err != nil {
		return err
	}
	err = a.c.Login(a.cfg.Imap.User, p)
	return err
}

func (a *Account) terminate() error {
	if a.c != nil {
		return a.c.Logout()
	}
	return nil
}

func (a *Account) Run(responses chan<- workers.Message) {
	defer a.terminate()
	postResponse := func(msg workers.Message, id int) {
		msg.SetId(id)
		responses <- msg
	}
	for msg := range a.requests {
		switch msg := msg.(type) {
		case *workers.FetchMailboxes:
			result, err := a.handleFetchMailboxes(msg)
			var r workers.Message
			if err != nil {
				a.logger.Warnf("error fetching mailboxes %v", err)
				r = &workers.Error{Error: errors.New("error requesting imap server")}
			} else {
				r = &workers.MsgToDb{Wrapped: &workers.FetchMailboxesImapRes{
					Mailboxes:result,
				}}
			}
			postResponse(r, msg.GetId())
		case *workers.ConnectImap:
			var r workers.Message
			if err := a.connect(); err != nil {
				a.logger.Warnf("error connecting to imap server %v", err)
				r = &workers.Error{Error: err}
			} else {
				r = &workers.Done{}
			}
			postResponse(r, msg.GetId())
		}
	}
}

func (a *Account) handleFetchMailboxes(msg *workers.FetchMailboxes) ([]*models.Mailbox, error) {
	result := make([]*models.Mailbox, 0)
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- a.c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		result = append(result, &models.Mailbox{m.Name})
	}

	if err := <-done; err != nil {
		return nil, err
	}
	return result, nil
}
