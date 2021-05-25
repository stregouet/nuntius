package migrations

import (
	"github.com/stregouet/nuntius/database"
)

func init() {
	database.Register(&database.Migration{
		Version:     "20210426",
		Description: "initial schema",
		Statements: []string{
			`CREATE TABLE mail (
				id INTEGER PRIMARY KEY,
				threadid INTEGER,
				date datetime,
				uid INTEGER,
				subject TEXT,
				messageid TEXT,
				inreplyto TEXT
			)`,
			"CREATE INDEX subject_idx ON mail(subject)",
			"CREATE INDEX messageid_idx ON mail(messageid)",
			"CREATE INDEX inreplyto_idx ON mail(inreplyto)",
			"CREATE TABLE counter (name text, value integer)",
			"INSERT INTO counter (name, value) VALUES ('threadid', 0)",
		},
	})
}
