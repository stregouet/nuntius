package config

import (
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
    Accounts []Account
}

func (c *Config) Validate() error {
    _, err := lib.LogParseLevel(c.Log.Level)
    if err != nil {
        return err
    }
    return nil
}
