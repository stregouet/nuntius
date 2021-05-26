package main

import (
	"log"

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
	l.Debugf("config %#v", c.Accounts[0].Imap)
	if err = ui.InitApp(l, c); err != nil {
		l.Fatalf("cannot init app %v", err)
	}

	ui.App.Run()
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
	return &c, nil
}
