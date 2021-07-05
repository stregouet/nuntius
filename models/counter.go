package models

import (
	ndb "github.com/stregouet/nuntius/database"
)

type ThreadCounter struct {
	id int
}

func FetchThreadCounter(r ndb.Queryer) (ThreadCounter, error) {
	var id ThreadCounter
	err := r.QueryRow("select value from counter where name = 'threadid'").Scan(&id.id)
	return id, err
}

func (t *ThreadCounter) Next() int {
	t.id++
	return t.id
}

func (t *ThreadCounter) UpdateThreadCounter(r ndb.Execer) error {
	_, err := r.Exec("update counter set value = ? where name = 'threadid'", t.id)
	return err
}
