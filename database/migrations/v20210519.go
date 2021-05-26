package migrations

import (
	"github.com/stregouet/nuntius/database"
)

func init() {
	database.Register(&database.Migration{
		Version:     "20210519",
		Description: "account relation",
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
			"ALTER TABLE mail ADD COLUMN account INTEGER NOT NULL REFERENCES account(id) ON DELETE CASCADE",
			"ALTER TABLE mail ADD COLUMN mailbox INTEGER NOT NULL REFERENCES mailbox(id) ON DELETE CASCADE",
		},
	})
}
