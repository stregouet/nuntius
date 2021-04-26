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

func (m *Mail) UpdateThreadid(tx *sql.Tx) error {
  threadid, err := m.hasParent(tx)
  if err != nil {
    return errors.Wrap(err, "while checking parent")
  }
  if threadid != 0 {
    // we find a parent with threadid
    m.Threadid = threadid
    _, err = tx.Exec("UPDATE mail SET threadid = ? WHERE inreplyto = ?", threadid, m.MessageId)
    if err != nil {
      return errors.Wrap(err, "while updating children threadid")
    }
  } else {
    threadid, err = m.hasChild(tx)
    if err != nil {
      return errors.Wrap(err, "while checking child")
    }
    if threadid != 0 {
      // we find child with threadid
      m.Threadid = threadid
      _, err = tx.Exec("UPDATE mail SET threadid = ? WHERE inreplyto = ?", threadid, m.MessageId)
      if err != nil {
        return errors.Wrap(err, "while updating children threadid")
      }
    } else {
      row := tx.QueryRow("UPDATE counter SET value = value + 1 WHERE name = 'threadid' RETURNING value")
      var id int
      err = row.Scan(&id)
      if err != nil {
        return errors.Wrap(err, "while updating `threadid` counter")
      }
      m.Threadid = id
    }
  }

  return nil
}
