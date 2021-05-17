package workers

import (
	"database/sql"
	"errors"
	"log"
)

type Database struct {
	*BaseWorker

	logger *log.Logger
}

func NewDatabase(l *log.Logger) *Database {
	bw := &BaseWorker{
		requests:  make(chan Message, 10),
		responses: make(chan Message, 10),
	}
	return &Database{bw, l}
}

func (d *Database) Run() {
	db, err := sql.Open("sqlite3", "sqlite.db")
	if err != nil {
		d.logger.Fatalf("cannot open db %v", err)
	}
	defer db.Close()
	for {
		select {
		case msg := <-d.requests:
			d.handleMessage(db, msg)
		}
	}
}

func (d *Database) handleMessage(db *sql.DB, msg Message) {
	switch msg := msg.(type) {
	case *FetchMailbox:
		result, err := d.handleFetchMailbox(db, msg.Mailbox)
		if err != nil {
			d.responses <- &Error{Error: errors.New("oups fetch mailbox")}
		}
		d.responses <- WithId(
			&FetchMailboxRes{List: result},
			msg.GetId())
	}

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
