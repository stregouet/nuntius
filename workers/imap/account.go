package imap

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-message/textproto"
	"github.com/pkg/errors"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/workers"
)

var PARSE_IMAP_ERR = errors.New("error parsing mail from imap server")

type Account struct {
	cfg          *config.Account
	requests     chan workers.Message
	c            *client.Client
	selectedMbox *models.Mailbox
	logger       *lib.Logger
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
func (a *Account) selectMbox(mailbox string) error {
	if a.selectedMbox != nil && a.selectedMbox.Name == mailbox {
		return nil
	}
	res, err := a.c.Select(mailbox, false)
	if err != nil {
		return err
	}
	a.selectedMbox = &models.Mailbox{
		Name:     mailbox,
		ReadOnly: res.ReadOnly,
		Count:    res.Messages,
		Unseen:   res.Unseen,
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
					Mailboxes: result,
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
		case *workers.FetchMailbox:
			result, err := a.handleFetchMailbox(msg.Mailbox)
			var r workers.Message
			if err != nil {
				a.logger.Warnf("error fetching mailbox %v", err)
				r = &workers.Error{Error: errors.New("error requesting imap server")}
			} else {
				r = &workers.MsgToDb{Wrapped: &workers.FetchMailboxImapRes{
					Mailbox: msg.Mailbox,
					Mails:   result,
				}}
			}
			postResponse(r, msg.GetId())
		}
	}
}

func (a *Account) handleFetchMailbox(mailbox string) ([]*models.Mail, error) {
	err := a.selectMbox(mailbox)
	if err != nil {
		return nil, err
	}
	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Now().Add(-48 * time.Hour)
	uids, err := a.c.UidSearch(criteria)
	if err != nil {
		return nil, err
	}
	section := &imap.BodySectionName{
		BodyPartName: imap.BodyPartName{
			Specifier: imap.HeaderSpecifier,
		},
		Peek: true,
	}

	items := []imap.FetchItem{
		imap.FetchBodyStructure,
		imap.FetchEnvelope,
		imap.FetchInternalDate,
		imap.FetchFlags,
		imap.FetchUid,
		section.FetchItem(),
	}
	result := make([]*models.Mail, 0)
	err = fetch(a.c, uids, items, func(m *imap.Message) error {
		reader := m.GetBody(section)
		textprotoHeader, err := textproto.ReadHeader(bufio.NewReader(reader))
		if err != nil {
			return fmt.Errorf("could not read header: %v", err)
		}
		header := &mail.Header{message.Header{textprotoHeader}}

		inreplies, err := header.MsgIDList("in-reply-to")
		if err != nil {
			a.logger.Errorf("cannot parse in-reply, %v", err)
			return PARSE_IMAP_ERR
		}
		inreply := ""
		if len(inreplies) > 0 {
			inreply = inreplies[0]
		}
		msgid, err := header.MessageID()
		if err != nil {
			a.logger.Errorf("cannot parse messageid, %v", err)
			return PARSE_IMAP_ERR
		}

		mail := &models.Mail{
			Subject:   m.Envelope.Subject,
			InReplyTo: inreply,
			MessageId: msgid,
			Date:      m.Envelope.Date,
			Flags:     m.Flags,
			Uid:       m.Uid,
			Header:    header,
		}
		result = append(result, mail)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func canOpen(mbox *imap.MailboxInfo) bool {
	for _, attr := range mbox.Attributes {
		if attr == imap.NoSelectAttr {
			return false
		}
	}
	return true
}

func (a *Account) handleFetchMailboxes(msg *workers.FetchMailboxes) ([]*models.Mailbox, error) {
	result := make([]*models.Mailbox, 0)
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- a.c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		parent := ""
		if !canOpen(m) {
			continue
		}
		parts := strings.Split(m.Name, m.Delimiter)
		shortName := parts[len(parts)-1]
		if len(parts) > 1 {
			parent = parts[len(parts)-2]
		}
		mbox := models.Mailbox{Name: m.Name, ShortName: shortName, Parent: parent}
		result = append(result, &mbox)
	}

	if err := <-done; err != nil {
		return nil, err
	}
	return result, nil
}
