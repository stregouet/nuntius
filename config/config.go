package config

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/stregouet/nuntius/lib"
)

type ImapCfg struct {
	Port    uint16
	Host    string
	User    string
	Tls     bool
	PassCmd string
}

type SmtpCfg struct {
	Port    uint16
	Host    string
	User    string
	Tls     bool
	PassCmd string
	// either plain, login, none
	Auth string
}

type Account struct {
	Name string
	Imap *ImapCfg
	Smtp *SmtpCfg
}

type Filters map[string]string

type Config struct {
	Log struct {
		Level  string
		Output string
	}
	Accounts    []*Account
	Keybindings Keybindings
	Filters     map[string]string
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

func (c *Config) valideFiltersMime() error {
	for mime, _ := range c.Filters {
		parts := strings.Split(mime, "/")
		if len(parts) != 2 {
			return fmt.Errorf("malformed mime `%s`", mime)
		}
		if parts[0] == "*" {
			return fmt.Errorf("mime part should not contain `*`, only submime (%s)", mime)
		}
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
	if err = c.Keybindings.Validate(); err != nil {
		return err
	}
	if err = c.valideFiltersMime(); err != nil {
		return errors.Wrap(err, "in filters section")
	}
	return nil
}
