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
				name TEXT UNIQUE,
				account INTEGER NOT NULL REFERENCES account(id) ON DELETE CASCADE,
				UNIQUE (name, account)
			)`,
			"ALTER TABLE mail ADD COLUMN account INTEGER NOT NULL REFERENCES account(id) ON DELETE CASCADE",
			"ALTER TABLE mail ADD COLUMN mailbox INTEGER NOT NULL REFERENCES mailbox(id) ON DELETE CASCADE",
		},
	})
}
