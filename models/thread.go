package models

import (
	"database/sql"

	"github.com/pkg/errors"

	ndb "github.com/stregouet/nuntius/database"
)

type Thread struct {
	Id      int
	Subject string
	Date    string
}

func (m *Thread) ToRune() []rune {
	return []rune(m.Subject)
}

type Mail struct {
	Id        int
	Threadid  int
	Account   string
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

func (m *Mail) InsertInto(r ndb.Execer, mailbox, accname string) error {
	_, err := r.Exec(`INSERT INTO mail (subject, messageid, inreplyto, date, threadid, account, mailbox)
SELECT ?, ?, ?, ?, ?, account.id, mailbox.id
FROM
  mailbox
  JOIN account on account.id = mailbox.account
WHERE mailbox.name = ? AND account.name = ?`,
		m.Subject,
		m.MessageId,
		m.InReplyTo,
		m.Date,
		m.Threadid,
		mailbox,
		accname,
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
