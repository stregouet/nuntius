package config

import (
	"fmt"
	"strings"

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

var KEYS_MODES = []KeyMode{KEY_MODE_SEARCH, KEY_MODE_THREAD, KEY_MODE_GLOBAL, KEY_MODE_MBOXES}

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
	modes := make([]string, 0, len(KEYS_MODES))
	for _, m := range KEYS_MODES {
		modes = append(modes, string(m))
	}
	return fmt.Errorf("unknown keymode `%s` (available modes: %s)", k, strings.Join(modes, ", "))
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
