package migrations

import (
	"github.com/stregouet/nuntius/database"
)

func init() {
	database.Register(&database.Migration{
		Version:     "20210426",
		Description: "initial schema",
		Statements: []string{
			`CREATE TABLE account (
				id INTEGER PRIMARY KEY,
				name TEXT UNIQUE
			)`,
			`CREATE TABLE mailbox (
				id INTEGER PRIMARY KEY,
				name TEXT,
				shortname TEXT,
				parent TEXT,
				account INTEGER NOT NULL REFERENCES account(id) ON DELETE CASCADE,
				UNIQUE (name, account)
			)`,
			"CREATE INDEX mailbox_parent_idx ON mailbox(parent)",
			`CREATE TABLE mail (
				id INTEGER PRIMARY KEY,
				threadid INTEGER,
				date datetime,
				uid INTEGER,
				flags TEXT,
				subject TEXT,
				messageid TEXT,
				inreplyto TEXT,
				mailbox INTEGER NOT NULL REFERENCES mailbox(id) ON DELETE CASCADE,
				account INTEGER NOT NULL REFERENCES account(id) ON DELETE CASCADE,
				UNIQUE (uid, mailbox)
			)`,
			"CREATE INDEX subject_idx ON mail(subject)",
			"CREATE INDEX messageid_idx ON mail(messageid)",
			"CREATE INDEX inreplyto_idx ON mail(inreplyto)",
			"CREATE TABLE counter (name text, value integer)",
			"INSERT INTO counter (name, value) VALUES ('threadid', 0)",
		},
	})
}
