package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/pkg/errors"

	ndb "github.com/stregouet/nuntius/database"
)

type Thread struct {
	Id      int
	Subject string
	Date    time.Time
	Count   int
}

func (m *Thread) ToRune() []rune {
	return []rune(fmt.Sprintf("%s (%d) %s",
		m.Date.Format("2006-01-02 15:04:05"),
		m.Count,
		m.Subject,
	))
}

type Mail struct {
	Id        int
	Uid       uint32
	Threadid  int
	Subject   string
	Flags     []string
	MessageId string
	Mailbox   string
	InReplyTo string
	Date      time.Time
	Header    *mail.Header
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
	_, err := r.Exec(`INSERT INTO mail (subject, messageid, inreplyto, date, threadid, uid, flags, account, mailbox)
SELECT ?, ?, ?, ?, ?, ?, ?, account.id, mailbox.id
FROM
  mailbox
  JOIN account on account.id = mailbox.account
WHERE mailbox.name = ? AND account.name = ?
ON CONFLICT (uid, mailbox) DO UPDATE SET flags=excluded.flags`,
		m.Subject,
		m.MessageId,
		m.InReplyTo,
		m.Date,
		m.Threadid,
		m.Uid,
		strings.Join(m.Flags, ","),
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

func AllThreads(r ndb.Queryer, mailbox, accname string) ([]*Thread, error) {
	// select all threads in specified account, mailbox with:
	// - count of messages in this thread
	// - date of the most recent messages in this thread
	// - subject of root of this thread
	rows, err := r.Query(`WITH q AS (
  SELECT
    subject,
    threadid,
	inreplyto,
    COUNT(1) OVER w as count,
    MAX(date) OVER w as date
  FROM
    mail m
    JOIN mailbox mbox ON mbox.id = m.mailbox
    JOIN account a ON a.id = m.account AND a.id = mbox.account
  WHERE
    a.name = ? AND
    mbox.name = ?
  WINDOW w AS (partition by threadid)
) SELECT
  q.threadid, q.subject, q.date, q.count
FROM q
WHERE
  NOT EXISTS (SELECT 1 FROM mail p WHERE q.inreplyto = p.messageid)
`, accname, mailbox)
	if err != nil {
		return nil, err
	}
	result := make([]*Thread, 0)
	for rows.Next() {
		var threadid sql.NullInt32
		var subject string
		var date DateFromStr
		var count int
		err = rows.Scan(&threadid, &subject, &date, &count)
		if err != nil {
			return nil, err
		}
		t := &Thread{Subject: subject, Date: date.T, Count: count}
		if threadid.Valid {
			t.Id = int(threadid.Int32)
		}
		result = append(result, t)
	}
	return result, nil
}

func AllThreadsRoot(r ndb.Queryer, accname string) ([]*Mail, error) {
	rows, err := r.Query(`
SELECT m.id
FROM
  mail m
  JOIN account a on m.account = a.id
WHERE
  a.name = ? AND
  NOT EXISTS (SELECT 1 FROM mail p WHERE m.inreplyto = p.messageid)`, accname)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roots := make([]*Mail, 0)
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		roots = append(roots, &Mail{Id: id})
	}
	return roots, nil
}
