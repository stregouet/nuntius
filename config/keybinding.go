package config

import (
	"fmt"

	"github.com/stregouet/nuntius/lib"
)

type Command string
type KeyMode string

// map key sequence to command
type Mapping map[string]Command
type Keybindings map[KeyMode]Mapping

const (
	KEY_MODE_SEARCH KeyMode = "search"
	KEY_MODE_THREAD KeyMode = "thread"
	KEY_MODE_GLOBAL KeyMode = "global"
	KEY_MODE_MBOXES KeyMode = "mboxes"
)

var KEYS_MODES = []KeyMode{KEY_MODE_SEARCH, KEY_MODE_THREAD, KEY_MODE_GLOBAL}

func (m Mapping) FindCommand(ks []*lib.KeyStroke) string {
	s := lib.KeyStrokesToString(ks)
	return string(m[s])
}

func (k KeyMode) Validate() error {
	for _, mode := range KEYS_MODES {
		if mode == k {
			return nil
		}
	}
	return fmt.Errorf("unknown keymode `%s` (available modes: XXX)", k)
}

func (k Keybindings) Validate() error {
	for kmode, _ := range k {
		if err := kmode.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (k Keybindings) Defaults() {
	for _, mode := range KEYS_MODES {
		if _, ok := k[mode]; !ok {
			k[mode] = make(Mapping)
		}
	}
}