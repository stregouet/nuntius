package workers

import (
	"database/sql"

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
	case *FetchMailboxImapRes:
		result, err := d.handleFetchMailboxImap(db, msg)
		var m Message
		if err != nil {
			m = &Error{Error: errors.New("oups inserting mails")}
			d.logger.Errorf("error while inserting and fetching mails %v", err)
		} else {
			m = &FetchMailboxRes{List: result}
		}
		d.postResponse(m, msg.GetId())
	case *FetchMailbox:
		result, err := d.handleFetchMailbox(db, msg)
		var m Message
		if err != nil {
			m = &Error{Error: errors.New("oups fetch mailbox")}
			d.logger.Errorf("error while fetchingmailbox %v", err)
		} else {
			m = &FetchMailboxRes{List: result}
		}
		d.postResponse(m, msg.GetId())
	}
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


func (d *Database) handleFetchMailboxImap(db *sql.DB, msg *FetchMailboxImapRes) ([]*models.Thread, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning tx")
	}
	for _, m := range msg.Mails {
		err = m.InsertInto(tx, msg.Mailbox, msg.GetAccName())
		if err != nil {
			if rollerr := tx.Rollback(); rollerr != nil {
				return nil, errors.Wrap(rollerr, "while trying to rollback")
			}
			return nil, errors.Wrapf(err, "while inserting mail (m: %#v)", m)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting tx")
	}
	return models.AllThreads(db, msg.Mailbox, msg.GetAccName())
}

func (d *Database) handleFetchMailbox(db *sql.DB, msg *FetchMailbox) ([]*models.Thread, error) {
	return models.AllThreads(db, msg.Mailbox, msg.GetAccName())
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
