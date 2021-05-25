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
		return errors.Wrap(err, "while beginning transaction")
	}
	err = ndb.Setup(tx)
	if err != nil {
		return errors.Wrap(err, "while creating _migrations table")
	}
	err = ndb.Migrate(tx)
	if err != nil {
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
	d.logger.Debugf("db will post response message id:%d, (type %#v)", msg.GetId(), msg.(*FetchMailboxesRes))
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
			m = &Error{Error: errors.New("oups fetch mailbox")}
			d.logger.Errorf("error while fetchingmailboxes %v", err)
		} else {
			m = &FetchMailboxesRes{Mailboxes: result}
		}
		// tmp := m.SetId(msg.GetId())
		// d.responses <- m.SetId(msg.GetId())
		d.postResponse(m, msg.GetId())
	case *FetchMailbox:
		result, err := d.handleFetchMailbox(db, msg.Mailbox)
		var m Message
		if err != nil {
			m = &Error{Error: errors.New("oups fetch mailbox")}
		} else {
			m = &FetchMailboxRes{List: result}
		}
		d.postResponse(m, msg.GetId())
	}
}

// return id for account with name `accountname`, in case this account does not exist
// it will insert it
func (d *Database) getOrCreateAccount(db *sql.DB, accountname string) (int, error) {
	row := db.QueryRow("select id from account where name = ?", accountname)
	var id int
	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		// no account with this name => we should create one
		result, err := db.Exec("insert into account (name) values (?) returning *", accountname)
		if err != nil {
			return 0, err
		}
		v, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(v), nil
	} else if err != nil {
		return 0, err
	} else {
		return id, nil
	}
}


func (d *Database) handleFetchMailboxes(db *sql.DB, accountname string) ([]*models.Mailbox, error) {
	accid, err := d.getOrCreateAccount(db, accountname)
	if err != nil {
		return nil, errors.Wrap(err, "while getting account id")
	}
	rows, err := db.Query("select mailbox from mail where account = ?", accid)
	if err != nil {
		return nil, errors.Wrap(err, "while selecting mailboxes")
	}
	result := make([]*models.Mailbox, 0)
	for rows.Next() {
		var mailboxname string
		err = rows.Scan(&mailboxname)
		if err != nil {
			return nil, errors.Wrap(err, "while scanning mailboxes")
		}
		m := &models.Mailbox{mailboxname}
		result = append(result, m)

	}
	return result, nil
}

func (d *Database) handleFetchMailbox(db *sql.DB, mailbox string) ([]string, error) {
	res := make([]string, 0)
	rows, err := db.Query("select id, subject from mail order by date desc limit 40")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var subject string
		err = rows.Scan(&id, &subject)
		if err != nil {
			return nil, err
		}
		res = append(res, subject)
	}
	return res, nil
}

func (d *Database) handleFetchMailboxesImap(db *sql.DB, msg *FetchMailboxesImapRes) ([]*models.Mailbox, error) {
	accid, err := d.getOrCreateAccount(db, msg.GetAccName())
	if err != nil {
		return nil, errors.Wrap(err, "while getting account id")
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning tx")
	}
	for _, m := range msg.Mailboxes {
		_, err = tx.Exec("insert into mailbox (name, account)  values (?, ?) on conflict (name, account) do nothing", m.Name, accid)
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
	return msg.Mailboxes, nil
}
