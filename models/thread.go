package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"

	ndb "github.com/stregouet/nuntius/database"
	"github.com/stregouet/nuntius/widgets"
)

// XXX rename to ThreadInfo?
type Thread struct {
	Id        int
	RootId    int
	Subject   string
	Date      time.Time
	Count     int
	HasUnread bool
}

func (t *Thread) StyledContent() []*widgets.ContentWithStyle {
	s := tcell.StyleDefault.Bold(t.HasUnread)
	return []*widgets.ContentWithStyle{
		{
			fmt.Sprintf("%s (%d) %s",
				t.Date.Format("2006-01-02 15:04:05"),
				t.Count,
				t.Subject),
			s,
		},
	}
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
	parts := []byte("[]")
	var err error
	if m.Parts != nil {
		parts, err = json.Marshal(m.Parts)
		if err != nil {
			return err
		}
	}
	// insert null instead of 0 for "empty" Threadid
	var threadid interface{}
	if m.Threadid != 0 {
		threadid = m.Threadid
	}
	// insert null instead of "" for "empty" InReplyTo
	var inreplyto interface{}
	if m.InReplyTo != "" {
		inreplyto = m.InReplyTo
	}
	// fake unique messageid instead of empty
	if m.MessageId == "" {
		m.MessageId = fmt.Sprintf("empty-%s-%s-%d", accname, mailbox, m.Uid)
	}
	res, err := r.Exec(`INSERT INTO mail (subject, messageid, inreplyto, date, threadid, uid, flags, parts, account, mailbox)
SELECT ?, ?, ?, ?, ?, ?, ?, ?, account.id, mailbox.id
FROM
  mailbox
  JOIN account on account.id = mailbox.account
WHERE mailbox.name = ? AND account.name = ?
ON CONFLICT (uid, mailbox) DO UPDATE SET flags=excluded.flags
ON CONFLICT (messageid) DO UPDATE SET identical_as=trim(printf('%s|(%s, %s)', mail.identical_as, excluded.uid, excluded.mailbox), '|')`,
		m.Subject,
		m.MessageId,
		inreplyto,
		m.Date,
		threadid,
		m.Uid,
		strings.Join(m.Flags, ","),
		parts,
		mailbox,
		accname,
	)
	if err != nil {
		return err
	}
	lastid, err := res.LastInsertId()
	if err != nil {
		return err
	}
	m.Id = int(lastid)
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

// fetch thread root
func (m *Mail) FetchRoot(r ndb.Queryer) (*Mail, error) {
	if strings.TrimSpace(m.MessageId) == "" {
		// no need to search for hierarchy in case of empty messageid
		return m, nil
	}
	var id int
	// retrieve parents in thread hierarchy, then select oldest parent
	err := r.QueryRow(`
WITH RECURSIVE empdata(id, inreplyto, messageid, date) as (
    SELECT
      mail.id,
      mail.inreplyto,
      messageid,
	  mail.date
    FROM mail
    WHERE mail.id = ?
  UNION ALL
    SELECT
      this.id,
      this.inreplyto,
      this.messageid,
	  this.date
    FROM
      empdata prior
      INNER JOIN mail this ON this.messageid = prior.inreplyto
)
SELECT
  e.id
FROM
  (SELECT MIN(date) AS d FROM empdata) r,
  empdata e
WHERE e.date = r.d
	`, m.Id).Scan(&id)
	return &Mail{Id: id}, err
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
	// - subject of root of this thread (i.e. the oldest message)
	rows, err := r.Query(`
SELECT id, threadid, subject, mostrecent, seen, count
FROM (
    SELECT
	  p.id,
      p.threadid,
      subject,
	  MIN(flags like '%Seen%') OVER w AS seen,
      MAX(p.date) OVER w AS mostrecent,
      COUNT(1) OVER w as count,
      ROW_NUMBER() OVER (PARTITION BY threadid ORDER BY p.date ASC) AS rn
    FROM mail p
    WHERE p.threadid in (
      SELECT
        m.threadid
      FROM
        mail m
        JOIN mailbox mbox ON mbox.id = m.mailbox
        JOIN account a ON a.id = m.account AND a.id = mbox.account
      WHERE
        a.name = ? AND mbox.name = ?
    )
    WINDOW w AS (partition by threadid)
)
WHERE rn = 1
ORDER BY mostrecent DESC
`, accname, mailbox)
	if err != nil {
		return nil, err
	}
	result := make([]*Thread, 0)
	for rows.Next() {
		var rootid int
		var threadid sql.NullInt32
		var subject string
		var date DateFromStr
		var count int
		var seen bool
		err = rows.Scan(&rootid, &threadid, &subject, &date, &seen, &count)
		if err != nil {
			return nil, err
		}
		t := &Thread{RootId: rootid, Subject: subject, Date: date.T, Count: count, HasUnread: !seen}
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

// given the id of a mail this function returns all of its children in its thread
// with specific depth for each
func AllThreadMails(r ndb.Queryer, rootMailId int) ([]*Mail, error) {
	rows, err := r.Query(`
WITH RECURSIVE tmp(id, messageid, subject, date, uid, parts, flags, depth) as (
    SELECT
      mail.id,
      messageid,
	  subject,
	  date,
	  uid,
	  parts,
	  flags,
	  0 as depth
    FROM mail
    WHERE mail.id = ?
  UNION ALL
    SELECT
      this.id,
      this.messageid,
	  this.subject,
	  this.date,
	  this.uid,
	  this.parts,
	  this.flags,
	  prior.depth + 1 as depth
    FROM
      tmp prior
      INNER JOIN mail this ON this.inreplyto = prior.messageid
) select id, subject, date, uid, parts, flags, depth from tmp`, rootMailId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roots := make([]*Mail, 0)
	for rows.Next() {
		var id int
		var subject string
		var date time.Time
		var uid uint32
		var depth int
		var rawparts []byte
		var flags string
		err = rows.Scan(&id, &subject, &date, &uid, &rawparts, &flags, &depth)
		if err != nil {
			return nil, err
		}
		var parts []*BodyPart
		err = json.Unmarshal(rawparts, &parts)
		if err != nil {
			return nil, err
		}
		m := &Mail{Id: id, Subject: subject, depth: depth, Date: date, Uid: uid, Parts: parts}
		if flags != "" {
			// split only if flags is not empty
			// if flags is empty we want an empty []string
			m.Flags = strings.Split(flags, ",")
		}
		roots = append(roots, m)
	}
	return roots, nil
}
