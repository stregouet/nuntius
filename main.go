package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/stregouet/nuntius/config"
	_ "github.com/stregouet/nuntius/database/migrations"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/ui"
)

func main() {
	c, err := initConfig("config.toml")
	if err != nil {
		log.Fatalf("cannot initialize config `%v`", err)
	}
	l, err := lib.NewLogger(c.Log.Level, c.Log.Output)
	if err != nil {
		log.Fatal(err)
	}
	l.Debugf("config %#v", c.Keybindings)
	if err = ui.InitApp(l, c); err != nil {
		l.Fatalf("cannot init app %v", err)
	}

	defer recoverTerm(l) // recover upon panic and try restoring the pty
	ui.App.Run()
}

// recoverTerm prints the stacktrace upon panic and tries to recover the term
// not doing that leaves the terminal in a broken state
func recoverTerm(logger *lib.Logger) {
	var err interface{}
	logger.Debug("recoverterm called")
	if err = recover(); err == nil {
		logger.Debug("recoverterm no error")
		return
	}
	logger.Debugf("recoverterm error %v", err)
	debug.PrintStack()
	if ui.App != nil {
		ui.App.Close()
	}
	fmt.Fprintf(os.Stderr, "nuntius crashed: %v\n", err)
	os.Exit(1)
}

func initConfig(cfgFile string) (*config.Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "while reading config file")
	}

	var c config.Config
	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding config")
	}
	err = c.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "while validating config")
	}
	c.Keybindings.Defaults()
	return &c, nil
}
