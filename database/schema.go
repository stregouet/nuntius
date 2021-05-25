package database

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"sort"

	"github.com/pkg/errors"
)

var migrations map[string]*Migration = make(map[string]*Migration)

type Migration struct {
	Version     string
	Description string
	Statements  []string
	checksum    string
}

func (m *Migration) insertInto(tx *sql.Tx) error {
	_, err := tx.Exec(`
  INSERT INTO
    _migrations (version, description, sha1, installed_on) VALUES (
      ?, ?, ?, datetime('now')
  )`, m.Version, m.Description, m.checksum)
	return err
}

func Register(m *Migration) {
	h := sha1.New()
	for _, s := range m.Statements {
		io.WriteString(h, s)
	}
	m.checksum = hex.EncodeToString(h.Sum(nil))
	migrations[m.Version] = m
}

func Setup(tx *sql.Tx) error {
	_, err := tx.Exec("PRAGMA foreign_keys=on")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS _migrations(
    version TEXT PRIMARY KEY,
    description TEXT,
    sha1 TEXT,
    installed_on DATETIME
  )`)
	return err
}

func fetchMigrations(tx *sql.Tx) ([]map[string]string, error) {
	rows, err := tx.Query("SELECT version, sha1 FROM _migrations")
	if err != nil {
		return nil, errors.Wrap(err, "while selecting from _migrations")
	}
	defer rows.Close()
	result := make([]map[string]string, 0)
	for rows.Next() {
		var version string
		var checksum string
		err := rows.Scan(&version, &checksum)
		if err != nil {
			return nil, err
		}
		row := map[string]string{
			"version":  version,
			"checksum": checksum,
		}
		result = append(result, row)
	}
	return result, nil

}

func Migrate(tx *sql.Tx) error {
	rows, err := fetchMigrations(tx)
	if err != nil {
		return err
	}
	// hack: no set in golang, use a map instead
	doneVersion := make(map[string]struct{})
	for _, row := range rows {
		version := row["version"]
		if m, ok := migrations[version]; ok {
			// this migration already exist check sha1
			if m.checksum != row["checksum"] {
				return fmt.Errorf("bad sha1 for already existing migration `%s`", version)
			}
		} else {
			// unknown version !?!
			return fmt.Errorf("found unknown version in database (%s)", version)
		}
		doneVersion[version] = struct{}{}
	}

	versions := make([]string, 0, len(migrations))
	for version, _ := range migrations {
		versions = append(versions, version)
	}

	sort.Strings(versions)

	for _, version := range versions {
		if _, ok := doneVersion[version]; ok {
			// version already applied
			continue
		}
		migration := migrations[version]
		for _, statement := range migration.Statements {
			_, err = tx.Exec(statement)
			if err != nil {
				return errors.Wrapf(err, "while executing migration %s", version)
			}
		}
		if err = migration.insertInto(tx); err != nil {
			return err
		}
	}
	return nil
}
