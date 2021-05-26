package models


import (
    "errors"
    "time"
)

const DATE_SQLITE_FORMAT = "2006-01-02 15:04:05.999999999Z07:00"

type DateFromStr struct {
    T time.Time
}

func (d *DateFromStr) Scan(src interface{}) error {
    srcstr, ok := src.(string)
    if !ok {
        return errors.New("expect src to be string")
    }
    var err error
    d.T, err = time.Parse(DATE_SQLITE_FORMAT, srcstr)
    return err
}
