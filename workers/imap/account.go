package imap

import (
	"io"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
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
		case *workers.FetchNewMessages:
			r, err := a.handleFetchNewMessages(msg)
			if err != nil {
				a.logger.Warnf("error fetching new messages %v", err)
				r = &workers.Error{Error: errors.New("error requesting imap server")}
			}
			postResponse(r, msg.GetId())
		case *workers.FetchFullMail:
			r, err := a.handleFetchFullMail(msg)
			if err != nil {
				a.logger.Warnf("error fetching full message %v", err)
				r = &workers.Error{Error: errors.New("error requesting imap server")}
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

func (a *Account) handleFetchFullMail(msg *workers.FetchFullMail) (workers.Message, error) {
	err := a.selectMbox(msg.Mailbox)
	if err != nil {
		return nil, err
	}
	section := &imap.BodySectionName{Peek: true}
	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchFlags,
		imap.FetchUid,
		section.FetchItem(),
	}
	r := &workers.FetchFullMailRes{
		Filepath: fmt.Sprintf("/tmp/nuntius/%s/%d.mail", msg.Mailbox, msg.Uid),
	}
	if _, err := os.Stat(r.Filepath); os.IsNotExist(err) {
		if _, err := os.Stat(path.Dir(r.Filepath)); os.IsNotExist(err) {
			if err := os.MkdirAll(path.Dir(r.Filepath), 0755); err != nil {
				return nil, err
			}
		}
	} else {
		return r, nil
	}
	err = fetch(a.c, toSeqSet([]uint32{msg.Uid}), items, func(m *imap.Message) error {
		body := m.GetBody(section)
		if body == nil {
			return fmt.Errorf("could not get section %#v", section)
		}
		f, err := os.OpenFile(r.Filepath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, body)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// fetch new messages following strategy described in rfc4549#section-4.3.1
func (a *Account) handleFetchNewMessages(msg *workers.FetchNewMessages) (workers.Message, error) {
	err := a.selectMbox(msg.Mailbox)
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
	var set imap.SeqSet
	set.AddRange(msg.LastSeenUid+1, 0)
	err = fetch(a.c, &set, items, func(m *imap.Message) error {
		if m.Uid <= msg.LastSeenUid {
			// already known messages, and here we are only interested in new
			// messages
			return nil
		}
		reader := m.GetBody(section)
		textprotoHeader, err := textproto.ReadHeader(bufio.NewReader(reader))
		if err != nil {
			return fmt.Errorf("could not read header: %v", err)
		}
		header := &mail.Header{message.Header{textprotoHeader}}

		inreply := ""
		inreplies, err := header.MsgIDList("in-reply-to")
		if err != nil {
			a.logger.Warnf(
				"cannot parse in-reply (id: %s, subject: %s, inreplyto: %s) => ignoring, %v",
				m.Envelope.MessageId,
				m.Envelope.Subject,
				m.Envelope.InReplyTo,
				err)
		} else if len(inreplies) > 0 {
			inreply = inreplies[0]
		}
		msgid, err := header.MessageID()
		if err != nil {
			a.logger.Errorf(
				"cannot parse messageid (id: %s, subject: %s, date: %s) %v",
				m.Envelope.MessageId,
				m.Envelope.Subject,
				m.InternalDate,
				err)
			return PARSE_IMAP_ERR
		}

		mail := &models.Mail{
			Subject:   m.Envelope.Subject,
			InReplyTo: inreply,
			MessageId: msgid,
			Date:      m.Envelope.Date,
			Flags:     m.Flags,
			Parts:     models.BodyPartsFromImap(m.BodyStructure),
			Uid:       m.Uid,
			Header:    header,
		}
		result = append(result, mail)
		return nil
	})
	if err != nil {
		return nil, err
	}
	r := &workers.FetchNewMessagesRes{
		Mailbox: msg.Mailbox,
		Mails:   result,
	}
	return r, nil
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
	err = fetch(a.c, toSeqSet(uids), items, func(m *imap.Message) error {
		reader := m.GetBody(section)
		textprotoHeader, err := textproto.ReadHeader(bufio.NewReader(reader))
		if err != nil {
			return fmt.Errorf("could not read header: %v", err)
		}
		header := &mail.Header{message.Header{textprotoHeader}}

		inreplies, err := header.MsgIDList("in-reply-to")
		if err != nil {
			a.logger.Errorf("cannot parse in-reply (id: %s), %v", m.Envelope.MessageId, err)
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
