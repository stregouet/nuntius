package config

import (
    "errors"

    "github.com/stregouet/nuntius/lib"
)

type ImapCfg struct {
    Port uint16
    Host string
    User string
    Tls bool
    PassCmd string
}

type Account struct {
    Name string
    Imap *ImapCfg
}

type Config struct {
    Log struct {
        Level string
        Output string
    }
    Accounts []*Account
}

func (c *Config) Validate() error {
    _, err := lib.LogParseLevel(c.Log.Level)
    if err != nil {
        return err
    }
    if len(c.Accounts) < 1 {
        return errors.New("need at least one account")
    }
    return nil
}
