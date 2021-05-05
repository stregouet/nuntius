package models

import (
	"database/sql"

	"github.com/pkg/errors"
)

type Mail struct {
	Id        int
	Threadid  int
	Subject   string
	MessageId string
	Mailbox   string
	InReplyTo string
	Date      string
}

func (m *Mail) hasParent(tx *sql.Tx) (int, error) {
	row := tx.QueryRow("SELECT threadid FROM mail WHERE messageid = ?", m.InReplyTo)
	var id int
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (m *Mail) hasChild(tx *sql.Tx) (int, error) {
	row := tx.QueryRow("SELECT threadid FROM mail WHERE inreplyto = ?", m.MessageId)
	var id int
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (m *Mail) InsertInto(tx *sql.Tx) error {
	_, err := tx.Exec(
		"INSERT INTO mail (subject, threadid, messageid, mailbox, inreplyto, date) VALUES (?, ?, ?, ?, ?, ?)",
		m.Subject,
		m.Threadid,
		m.MessageId,
		m.Mailbox,
		m.InReplyTo,
		m.Date,
	)
	return err
}

func (m *Mail) applyThreadidOnChildren(tx *sql.Tx) error {
	_, err := tx.Exec(`with recursive empdata(id, messageid) as (
    select
      mail.id,
      messageid
    from mail
    where mail.inreplyto = ?
  UNION ALL
    select
      this.id,
      this.messageid
    from
      empdata prior
      inner join mail this on this.inreplyto = prior.messageid
	) update mail set threadid = ? from empdata where empdata.id = mail.id`, m.MessageId, m.Threadid)
	return err
}

// suppose to be called on root of thread
func (m *Mail) UpdateThreadidOnChild(tx *sql.Tx, threadid int) error {
	_, err := tx.Exec(`with recursive empdata(id, messageid) as (
    select
      mail.id,
      messageid
    from mail
    where mail.id = ?
  UNION ALL
    select
      this.id,
      this.messageid
    from
      empdata prior
      inner join mail this on this.inreplyto = prior.messageid
	) update mail set threadid = ? from empdata where empdata.id = mail.id`, m.Id, threadid)
	return err
}

func (m *Mail) UpdateThreadid(tx *sql.Tx) error {
	parentThread, err := m.hasParent(tx)
	if err != nil {
		return errors.Wrap(err, "while checking parent")
	}
	if parentThread != 0 {
		// we find a parent with parentThread
		m.Threadid = parentThread
		err = m.applyThreadidOnChildren(tx)
		if err != nil {
			return errors.Wrap(err, "while updating children threadid")
		}
	}
	childThreadid, err := m.hasChild(tx)
	if err != nil {
		return errors.Wrap(err, "while checking child")
	}
	if childThreadid != 0 {
		// we find child with childThreadid
		m.Threadid = childThreadid
		err = m.applyThreadidOnChildren(tx)
		if err != nil {
			return errors.Wrap(err, "while updating children threadid")
		}
	}

	if m.Threadid == 0 {
		row := tx.QueryRow("UPDATE counter SET value = value + 1 WHERE name = 'threadid' RETURNING value")
		var id int
		err = row.Scan(&id)
		if err != nil {
			return errors.Wrap(err, "while updating `threadid` counter")
		}
		m.Threadid = id
	}

	return nil
}
