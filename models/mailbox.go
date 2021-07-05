package models

import (
	"fmt"
	"strings"

	ndb "github.com/stregouet/nuntius/database"
	"github.com/stregouet/nuntius/widgets"
)

type Mailbox struct {
	Name        string
	Parent      string
	ShortName   string
	Count       uint32
	Unseen      uint32
	ReadOnly    bool
	LastSeenUid uint32

	directoryDepth int
}

func (m *Mailbox) StyledContent() []*widgets.ContentWithStyle {
	return []*widgets.ContentWithStyle{
		widgets.NewContent(m.ShortName),
	}
}

func (m *Mailbox) Depth() int {
	return m.directoryDepth
}

func (m *Mailbox) TabTitle() string {
	name := m.Name
	if m.ShortName != "" {
		name = m.ShortName
	}
	if len(name) > 8 {
		name = name[:8] + "…"
	}
	return name
}

func (m *Mailbox) UpdateLastUid(r ndb.Execer, accname string) error {
	_, err := r.Exec(
		"UPDATE mailbox SET lastseenuid = ? FROM account WHERE account.name = ? AND mailbox.name = ?",
		m.LastSeenUid,
		accname,
		m.Name,
	)
	return err
}

func (m *Mailbox) InsertInto(r ndb.Execer, accname string) error {
	columns := []string{"name", "shortname"}
	values := []interface{}{m.Name, m.ShortName}
	if m.Parent != "" {
		columns = append(columns, "parent")
		values = append(values, m.Parent)
	}
	values = append(values, accname)
	query := fmt.Sprintf(
		"INSERT INTO mailbox (%s) SELECT %s account.id FROM account WHERE account.name = ? ON CONFLICT (name, account) DO NOTHING",
		strings.Join(append(columns, "account"), ","),
		strings.Repeat("?,", len(columns)),
	)
	_, err := r.Exec(
		query,
		values...,
	)
	return err
}

func GetMailbox(r ndb.Queryer, mboxname, accname string) (*Mailbox, error) {
	var name string
	var shortname string
	var lastseenuid int
	err := r.QueryRow(`
SELECT
  m.name, m.shortname, m.lastseenuid
FROM
  mailbox m
  JOIN account a ON m.account = a.id
WHERE a.name = ? AND m.name = ?`,
		accname,
		mboxname,
	).Scan(&name, &shortname, &lastseenuid)
	if err != nil {
		return nil, err
	}
	m := &Mailbox{Name: name, ShortName: shortname, LastSeenUid: uint32(lastseenuid)}
	return m, nil
}

func AllMailboxes(r ndb.Queryer, accname string) ([]*Mailbox, error) {
	rows, err := r.Query(`WITH RECURSIVE tmp(id, name, shortname, depth) as (
  SELECT m.id, m.name, m.shortname, 0 as depth
  FROM
    mailbox m
    JOIN account a on m.account = a.id
  WHERE a.name = ? AND m.parent IS NULL
UNION ALL
  SELECT this.id, this.name, this.shortname, prior.depth + 1 as depth
  FROM
    tmp prior
	INNER JOIN mailbox this on this.parent = prior.shortname
) SELECT name, shortname, depth from tmp order by name;
	`, accname)
	if err != nil {
		return nil, err
	}
	result := make([]*Mailbox, 0)
	for rows.Next() {
		var name string
		var shortname string
		var depth int
		err = rows.Scan(&name, &shortname, &depth)
		if err != nil {
			return nil, err
		}
		m := &Mailbox{Name: name, ShortName: shortname, directoryDepth: depth}
		result = append(result, m)
	}
	return result, nil
}
