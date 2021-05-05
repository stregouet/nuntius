package models

import (
	"database/sql"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stregouet/nuntius/database"
	_ "github.com/stregouet/nuntius/database/migrations"
)

func setupdb(t *testing.T) (*sql.DB, error) {
	tmpfile, err := ioutil.TempFile("", "nuntius-test-*.db")
	if err != nil {
		return nil, err
	}
	tmpfile.Close()
	t.Cleanup(func() {
		// TODO: need to remove db-journal file
		os.Remove(tmpfile.Name())
	})

	db, err := sql.Open("sqlite3", tmpfile.Name())
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	err = database.Setup(tx)
	if err != nil {
		return nil, err
	}
	err = database.Migrate(tx)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return db, nil
}


func createMail(messagedid, inreplyto string) *Mail {
	return &Mail{
		MessageId: messagedid,
		InReplyTo: inreplyto,
	}
}

func fetch(tx *sql.Tx) (map[int]map[string]interface{}, error) {
	rows, err := tx.Query("select id, threadid, messageid, subject from mail")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[int]map[string]interface{})
	for rows.Next() {
		m := Mail{}
		d := make(map[string]interface{})
		rows.Scan(&m.Id, &m.Threadid, &m.MessageId, &m.Subject)
		d["threadid"] = m.Threadid
		d["messageid"] = m.MessageId
		result[m.Id] = d
	}
	return result, nil
}

func TestThreadidAttribution(t *testing.T) {
	db, err := setupdb(t)
	if err != nil {
		t.Fatalf("cannot setup database %v", err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("cannot begin transaction %v", err)
	}
	insertMail := func(m *Mail) *Mail {
		err = m.UpdateThreadid(tx)
		if err != nil {
			t.Fatal(err)
		}
		err = m.InsertInto(tx)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	m5 := insertMail(&Mail{MessageId: "id5", InReplyTo: "id4"})
	if m5.Threadid != 1 {
		t.Errorf("unexpected threadid for mail 5")
	}

	m6 := insertMail(&Mail{MessageId: "id6", InReplyTo: "id4"})
	if m6.Threadid != 2 {
		t.Errorf("unexpected threadid for mail 6")
	}

	m4 := insertMail(&Mail{MessageId: "id4", InReplyTo: "id1"})
	if m4.Threadid != 1 {
		t.Errorf("unexpected threadid for mail 4")
	}

	m1 := insertMail(&Mail{MessageId: "id1"})
	if m1.Threadid != 1 {
		t.Errorf("unexpected threadid for mail 1")
	}

	mails, err := fetch(tx)
	if err != nil {
		t.Fatalf("error fetching mails %v", err)
	}
	expected := map[int]map[string]interface{}{
		1: map[string]interface{}{
			"threadid": 1,
			"messageid": "id5",
		},
		2: map[string]interface{}{
			"threadid": 1,
			"messageid": "id6",
		},
		3: map[string]interface{}{
			"threadid": 1,
			"messageid": "id4",
		},
		4: map[string]interface{}{
			"threadid": 1,
			"messageid": "id1",
		},
	}
	if !reflect.DeepEqual(expected, mails) {
		t.Errorf("unexpected database content %v", mails)
	}
}

func TestThreadidAttributionOrder1(t *testing.T) {
	db, err := setupdb(t)
	if err != nil {
		t.Fatalf("cannot setup database %v", err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("cannot begin transaction %v", err)
	}
	insertMail := func(m *Mail) *Mail {
		err = m.UpdateThreadid(tx)
		if err != nil {
			t.Fatal(err)
		}
		err = m.InsertInto(tx)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	
	insertMail(&Mail{MessageId: "id5", InReplyTo: "id4"})
	insertMail(&Mail{MessageId: "id6", InReplyTo: "id4"})
	insertMail(&Mail{MessageId: "id1", InReplyTo: ""})

	mails, err := fetch(tx)
	if err != nil {
		t.Fatalf("error fetching mails %v", err)
	}
	expected := map[int]map[string]interface{}{
		1: map[string]interface{}{
			"threadid": 1,
			"messageid": "id5",
		},
		2: map[string]interface{}{
			"threadid": 2,
			"messageid": "id6",
		},
		3: map[string]interface{}{
			"threadid": 3,
			"messageid": "id1",
		},
	}
	if !reflect.DeepEqual(expected, mails) {
		t.Errorf("unexpected database content %v", mails)
	}

	insertMail(&Mail{MessageId: "id4", InReplyTo: "id1"})
	mails, err = fetch(tx)
	if err != nil {
		t.Fatalf("error fetching mails %v", err)
	}
	// we expect to choose threadid from parent of id4 (i.e. threadid of id1)
	expected = map[int]map[string]interface{}{
		1: map[string]interface{}{
			"threadid": 3,
			"messageid": "id5",
		},
		2: map[string]interface{}{
			"threadid": 3,
			"messageid": "id6",
		},
		3: map[string]interface{}{
			"threadid": 3,
			"messageid": "id1",
		},
		4: map[string]interface{}{
			"threadid": 3,
			"messageid": "id4",
		},
	}
	if !reflect.DeepEqual(expected, mails) {
		t.Errorf("unexpected database content %v", mails)
	}
}

func TestThreadidAttributionOrder2(t *testing.T) {
	db, err := setupdb(t)
	if err != nil {
		t.Fatalf("cannot setup database %v", err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("cannot begin transaction %v", err)
	}
	insertMail := func(m *Mail) *Mail {
		err = m.UpdateThreadid(tx)
		if err != nil {
			t.Fatal(err)
		}
		err = m.InsertInto(tx)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	insertMail(&Mail{MessageId: "id1", InReplyTo: ""})
	insertMail(&Mail{MessageId: "id2", InReplyTo: "id9"})
	insertMail(&Mail{MessageId: "id3", InReplyTo: "id11"})
	insertMail(&Mail{MessageId: "id4", InReplyTo: "id3"})
	insertMail(&Mail{MessageId: "id5", InReplyTo: "id4"})
	insertMail(&Mail{MessageId: "id6", InReplyTo: "id5"})
	insertMail(&Mail{MessageId: "id7", InReplyTo: "id6"})
	insertMail(&Mail{MessageId: "id8", InReplyTo: "id1"})
	insertMail(&Mail{MessageId: "id9", InReplyTo: "id8"})
	insertMail(&Mail{MessageId: "id10", InReplyTo: "id9"})
	insertMail(&Mail{MessageId: "id11", InReplyTo: "id10"})
	insertMail(&Mail{MessageId: "id12", InReplyTo: "id7"})

	mails, err := fetch(tx)
	if err != nil {
		t.Fatalf("error fetching mails %v", err)
	}
	expected := map[int]map[string]interface{}{
		1: map[string]interface{}{
			"threadid": 1,
			"messageid": "id1",
		},
		2: map[string]interface{}{
			"threadid": 1,
			"messageid": "id2",
		},
		3: map[string]interface{}{
			"threadid": 1,
			"messageid": "id3",
		},
		4: map[string]interface{}{
			"threadid": 1,
			"messageid": "id4",
		},
		5: map[string]interface{}{
			"threadid": 1,
			"messageid": "id5",
		},
		6: map[string]interface{}{
			"threadid": 1,
			"messageid": "id6",
		},
		7: map[string]interface{}{
			"threadid": 1,
			"messageid": "id7",
		},
		8: map[string]interface{}{
			"threadid": 1,
			"messageid": "id8",
		},
		9: map[string]interface{}{
			"threadid": 1,
			"messageid": "id9",
		},
		10: map[string]interface{}{
			"threadid": 1,
			"messageid": "id10",
		},
		11: map[string]interface{}{
			"threadid": 1,
			"messageid": "id11",
		},
		12: map[string]interface{}{
			"threadid": 1,
			"messageid": "id12",
		},
	}
	if !reflect.DeepEqual(expected, mails) {
		t.Errorf("unexpected database content %v", mails)
	}
}
