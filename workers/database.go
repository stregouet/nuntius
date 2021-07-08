package workers

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	ndb "github.com/stregouet/nuntius/database"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
)

type Database struct {
	*BaseWorker

	logger *lib.Logger
}

func NewDatabase(l *lib.Logger) *Database {
	bw := &BaseWorker{
		requests:  make(chan Message, 10),
		responses: make(chan Message, 10),
	}
	return &Database{bw, l}
}

func (d *Database) setup(db *sql.DB) error {
	_, err := db.Exec("PRAGMA foreign_keys=on")
	if err != nil {
		return errors.Wrap(err, "while setting foreign_keys pragma")
	}
	tx, err := db.Begin()
	if err != nil {
		if rollerr := tx.Rollback(); rollerr != nil {
			return errors.Wrap(rollerr, "while trying to rollback")
		}
		return errors.Wrap(err, "while beginning transaction")
	}
	err = ndb.Setup(tx)
	if err != nil {
		if rollerr := tx.Rollback(); rollerr != nil {
			return errors.Wrap(rollerr, "while trying to rollback")
		}
		return errors.Wrap(err, "while creating _migrations table")
	}
	err = ndb.Migrate(tx)
	if err != nil {
		if rollerr := tx.Rollback(); rollerr != nil {
			return errors.Wrap(rollerr, "while trying to rollback")
		}
		return errors.Wrap(err, "while migrating db")
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "while commiting")
	}
	return nil
}

func (d *Database) Run() {
	defer lib.Recover(d.logger, nil)
	db, err := sql.Open("sqlite3", "sqlite.db")
	if err != nil {
		d.logger.Errorf("cannot open db %v", err)
		panic("cannot open db")
	}
	defer db.Close()
	err = d.setup(db)
	if err != nil {
		d.logger.Errorf("cannot setup db %v", err)
		panic("cannot setup db")
	}
	for {
		select {
		case msg := <-d.requests:
			d.handleMessage(db, msg)
		}
	}
}

func (d *Database) postResponse(msg Message, id int) {
	msg.SetId(id)
	d.responses <- msg
}

func (d *Database) handleMessage(db *sql.DB, msg Message) {
	switch msg := msg.(type) {
	case *FetchMailboxesImapRes:
		result, err := d.handleFetchMailboxesImap(db, msg)
		var m Message
		if err != nil {
			m = &Error{Error: errors.New("oups inserting mailboxes")}
			d.logger.Errorf("error while inserting mailboxes %v", err)
		} else {
			m = &FetchMailboxesRes{Mailboxes: result}
		}
		d.postResponse(m, msg.GetId())
	case *FetchMailboxes:
		result, err := d.handleFetchMailboxes(db, msg.GetAccName())
		var m Message
		if err != nil {
			m = &Error{Error: errors.New("oups fetch mailboxes")}
			d.logger.Errorf("error while fetchingmailboxes %v", err)
		} else {
			m = &FetchMailboxesRes{Mailboxes: result}
		}
		// tmp := m.SetId(msg.GetId())
		// d.responses <- m.SetId(msg.GetId())
		d.postResponse(m, msg.GetId())
	case *SaveMailFlags:
		r, err := d.handleSaveMailFlags(db, msg)
		if err != nil {
			r = &Error{Error: errors.New("cannot save flags")}
			d.logger.Errorf("whire updating flags %v (mailid: %d)", err, msg.MailId)
		}
		d.postResponse(r, msg.GetId())
	case *InsertNewMessages:
		result, err := d.handleInsertNewMessages(db, msg)
		var m Message
		if err != nil {
			m = &Error{Error: errors.New("oups inserting mails")}
			d.logger.Errorf("error while inserting new mails %v", err)
		} else {
			m = &InsertNewMessagesRes{Threads: result}
		}
		d.postResponse(m, msg.GetId())
	case *UpdateMessages:
		result, err := d.handleUpdateMessages(db, msg)
		var m Message
		if err != nil {
			m = &Error{Error: errors.New("cannot update mails flags")}
			d.logger.Errorf("error while updating mails flags %v", err)
		} else {
			m = &UpdateMessagesRes{Threads: result}
		}
		d.postResponse(m, msg.GetId())
	case *FetchThread:
		result, err := d.handleFetchThread(db, msg)
		var m Message
		if err != nil {
			m = &Error{Error: errors.New("oups fetching thread")}
			d.logger.Errorf("error while fetching mail %v", err)
		} else {
			m = &FetchThreadRes{Mails: result}
		}
		d.postResponse(m, msg.GetId())
	case *FetchMailbox:
		m, err := d.handleFetchMailbox(db, msg)
		if err != nil {
			m = &Error{Error: errors.New("oups fetch mailbox")}
			d.logger.Errorf("error while fetchingmailbox %v", err)
		}
		d.postResponse(m, msg.GetId())

	}
}

func (d *Database) handleInsertNewMessages(db *sql.DB, msg *InsertNewMessages) ([]*models.Thread, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning tx")
	}

	rollback := func(err error, msg string) error {
		if rollerr := tx.Rollback(); rollerr != nil {
			return errors.Wrap(rollerr, "while trying to rollback")
		}
		return errors.Wrap(err, msg)
	}
	// first insert mails on db
	lastuid := uint32(0)
	for _, m := range msg.Mails {
		err = m.InsertInto(tx, msg.Mailbox, msg.GetAccName())
		if err != nil {
			return nil, rollback(err, fmt.Sprintf("while inserting mail (m: %#v)", m))
		}
		if m.Uid > lastuid {
			lastuid = m.Uid
		}
	}
	// update lastseenuid for this mailbox
	m := models.Mailbox{Name: msg.Mailbox, LastSeenUid: lastuid}
	err = m.UpdateLastUid(tx, msg.GetAccName())
	if err != nil {
		return nil, rollback(
			err,
			fmt.Sprintf("while updating lastseenuid (mbox: %#v)", msg.Mailbox))
	}
	// then fetch last thread id
	threadid, err := models.FetchThreadCounter(tx)
	if err != nil {
		return nil, rollback(err, "while fetching thread counter")
	}
	// then insert new threadid
	alreadydone := make(map[int]struct{})
	for idx, m := range msg.Mails {
		root, err := m.FetchRoot(tx)
		if err != nil {
			return nil, rollback(err, fmt.Sprintf("while fetching root %d", m.Uid))
		}
		if _, ok := alreadydone[root.Id]; ok {
			continue
		}
		err = root.UpdateThreadidOnChild(tx, threadid.Next())
		if err != nil {
			return nil, rollback(err, "while updating threadid")
		}
		alreadydone[root.Id] = struct{}{}
		if idx > 0 && idx%100 == 0 {
			d.logger.Debugf("update root (%d/%d)", idx, len(msg.Mails))
		}
	}
	// finally update last threadid
	err = threadid.UpdateThreadCounter(tx)
	if err != nil {
		return nil, rollback(err, "while updating thread counter")
	}
	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting tx")
	}
	return models.AllThreads(db, msg.Mailbox, msg.GetAccName())
}

func (d *Database) handleFetchMailboxes(db *sql.DB, accountname string) ([]*models.Mailbox, error) {
	_, err := db.Exec("insert into account (name) values (?) on conflict (name) do nothing", accountname)
	if err != nil {
		return nil, errors.Wrap(err, "while inserting account")
	}
	rows, err := models.AllMailboxes(db, accountname)
	if err != nil {
		return nil, errors.Wrap(err, "while selecting mailboxes")
	}
	return rows, nil
}

func (d *Database) handleFetchThread(db *sql.DB, msg *FetchThread) ([]*models.Mail, error) {
	return models.AllThreadMails(db, msg.RootId)
}

func (d *Database) handleFetchMailbox(db *sql.DB, msg *FetchMailbox) (Message, error) {
	t, err := models.AllThreads(db, msg.Mailbox, msg.GetAccName())
	if err != nil {
		return nil, err
	}
	m, err := models.GetMailbox(db, msg.Mailbox, msg.GetAccName())
	if err != nil {
		return nil, err
	}
	return &FetchMailboxRes{List: t, LastSeenUid: m.LastSeenUid}, nil
}

func (d *Database) handleFetchMailboxesImap(db *sql.DB, msg *FetchMailboxesImapRes) ([]*models.Mailbox, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning tx")
	}
	for _, m := range msg.Mailboxes {
		err = m.InsertInto(tx, msg.GetAccName())
		if err != nil {
			if rollerr := tx.Rollback(); rollerr != nil {
				return nil, errors.Wrap(rollerr, "while trying to rollback")
			}
			return nil, errors.Wrap(err, "while inserting mailbox")
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting tx")
	}
	return d.handleFetchMailboxes(db, msg.GetAccName())
}

func (d *Database) handleUpdateMessages(db *sql.DB, msg *UpdateMessages) ([]*models.Thread, error) {
	// fetch mails in current mailbox ignoring mails with one greater than LastSeenUid
	// (such mails have been just inserted so no need to update them)
	dbMails, err := models.FetchMails(db, msg.Mailbox, msg.GetAccName(), msg.LastSeenUid)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching mails in mailbox")
	}
	imapMails := make(map[uint32]*models.Mail)
	for _, m := range msg.Mails {
		imapMails[m.Uid] = m
	}
	d.logger.Debugf("will update mails flags  (imap: %d, db: %d)", len(imapMails), len(dbMails))

	tx, err := db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning tx")
	}
	rollback := func(err error, msg string) error {
		if rollerr := tx.Rollback(); rollerr != nil {
			return errors.Wrap(rollerr, "while trying to rollback")
		}
		return errors.Wrap(err, msg)
	}
	for _, mail := range dbMails {
		if m, ok := imapMails[mail.Uid]; ok {
			err = mail.UpdateFlags(tx, m.Flags)
			if err != nil {
				return nil, rollback(err, "while updating flags")
			}
		} else {
			d.logger.Debugf("will delete mail %d", mail.Uid)
			err = mail.Delete(tx)
			if err != nil {
				return nil, rollback(err, "while deleting mail")
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting tx")
	}
	return models.AllThreads(db, msg.Mailbox, msg.GetAccName())
}

func (d *Database) handleSaveMailFlags(db *sql.DB, msg *SaveMailFlags) (Message, error) {
	m := &models.Mail{Id: msg.MailId, Flags: msg.Flags}
	err := m.SaveFlags(db)
	if err != nil {
		return nil, err
	}
	return &Done{}, nil
}
