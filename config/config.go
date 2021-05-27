package config

import (
    "errors"
    "fmt"

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

func (c *Config) uniqueAccountName() error {
	// hack: no set in golang, use a map instead
	names := make(map[string]struct{})
    for _, a := range c.Accounts {
		if _, ok := names[a.Name]; ok {
            return fmt.Errorf("account `%s` is defined twice", a.Name)
        }
        names[a.Name] = struct{}{}
    }
    return nil
}

func (c *Config) Validate() error {
    _, err := lib.LogParseLevel(c.Log.Level)
    if err != nil {
        return err
    }
    if len(c.Accounts) < 1 {
        return errors.New("need at least one account")
    }
    if err = c.uniqueAccountName(); err != nil {
        return err
    }
    return nil
}
