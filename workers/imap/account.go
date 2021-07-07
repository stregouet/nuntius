package imap

import (
	"io"
	"bufio"
	"fmt"
	"os"
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

func (a *Account) getImapPass() (string, error) {
	return getPass(a.cfg.Imap.PassCmd)
}

func (a *Account) getSmtpPass() (string, error) {
	return getPass(a.cfg.Smtp.PassCmd)
}

func (a *Account) connect() error {
	var err error
	if a.cfg.Imap.Tls {
		a.c, err = client.DialTLS(fmt.Sprintf("%s:%d", a.cfg.Imap.Host, a.cfg.Imap.Port), nil)
	}
	p, err := a.getImapPass()
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
	defer lib.Recover(a.logger, nil)
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
		case *workers.SendMail:
			var r workers.Message
			if err := a.handleSendMail(msg); err != nil {
				a.logger.Warnf("error sending mail %v", err)
				r = &workers.Error{Error: errors.New("error sending mail")}
			} else {
				r = &workers.Done{}
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
		case *workers.FetchMessageUpdates:
			r, err := a.handleFetchMessageUpdates(msg)
			if err != nil {
				a.logger.Warnf("error fetching messages update %v", err)
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
		}
	}
}

func (a *Account) handleFetchFullMail(msg *workers.FetchFullMail) (workers.Message, error) {
	err := a.selectMbox(msg.Mailbox)
	if err != nil {
		return nil, err
	}
	section := &imap.BodySectionName{Peek: true}  // XXX remove Peek
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

func (a *Account) handleFetchMessageUpdates(msg *workers.FetchMessageUpdates) (workers.Message, error) {
	a.logger.Debug("will fetch messages updates")
	err := a.selectMbox(msg.Mailbox)
	if err != nil {
		return nil, err
	}
	items := []imap.FetchItem{
		imap.FetchFlags,
		imap.FetchUid,
	}
	result := make([]*models.Mail, 0)
	var set imap.SeqSet
	// range 1:lastseenuid
	set.AddRange(1, msg.LastSeenUid)
	err = fetch(a.c, &set, items, func(m *imap.Message) error {
		mail := &models.Mail{
			Flags: m.Flags,
			Uid:   m.Uid,
		}
		result = append(result, mail)
		return nil
	})
	if err != nil {
		return nil, err
	}
	r := &workers.FetchMessageUpdatesRes{
		Mailbox: msg.Mailbox,
		Mails:   result,
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
	// range lastseenuid+1:*
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

		mail := &models.Mail{
			Subject:   m.Envelope.Subject,
			InReplyTo: m.Envelope.InReplyTo,
			MessageId: m.Envelope.MessageId,
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

func (a *Account) handleSendMail(msg *workers.SendMail) error {
	cfg := a.cfg.Smtp
	conn, err := connectSmtps(cfg.Host, cfg.Port)
	if err != nil {
		return err
	}
	defer conn.Close()
	password, err := a.getSmtpPass()
	if err != nil {
		return err
	}
	saslclient, err := newSaslClient(cfg.Auth, cfg.User, password)
	if err != nil {
		return err
	}
	if saslclient != nil {
		if err := conn.Auth(saslclient); err != nil {
			return errors.Wrap(err, "while issuing auth cmd")
		}
	}

	m, err := message.Read(msg.Body)
	if err != nil {
		return err
	}
	header := &mail.Header{message.Header{m.Header.Header}}
	err = encodeHeaderFields(header)
	if err != nil {
		return errors.Wrap(err, "while encoding header fields")
	}
	if !header.Has("Message-Id") {
		err := header.GenerateMessageID()
		if err != nil {
			return errors.Wrap(err, "generate message-id")
		}
	}
	if !header.Has("Date") {
		header.SetDate(time.Now())
	}

	from, err := header.AddressList("from")
	if err != nil {
		return errors.Wrapf(err, "addresslist `from` (%v)", header.Get("from"))
	}

	if err := conn.Mail(from[0].Address, nil); err != nil {
		return errors.Wrap(err, "while issuing mail cmd")
	}
	rcpts, err := listRecipients(header)
	if err != nil {
		return errors.Wrap(err, "addresslist rcpts")
	}
	for _, rcpt := range rcpts {
		if err := conn.Rcpt(rcpt.Address); err != nil {
			return errors.Wrap(err, "while issuing rcpt cmd")
		}
	}
	writer, err := conn.Data()
	if err != nil {
		return errors.Wrap(err, "while issuing data cmd")
	}
	defer writer.Close()


	header.SetContentType("text/plain", map[string]string{"charset": "UTF-8"})
	w, err := mail.CreateSingleInlineWriter(writer, *header)
	if err != nil {
		return errors.Wrap(err, "CreateSingleInlineWriter")
	}
	if _, err := io.Copy(w, m.Body); err != nil {
		return errors.Wrap(err, "io.Copy")
	}

	err = w.Close()
	if err != nil {
		return errors.Wrap(err, "will closing data writer")
	}

	conn.Quit()
	return nil
}
