package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	// "github.com/gdamore/tcell/v2"
	// "github.com/gdamore/tcell/v2/views"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	_ "github.com/mattn/go-sqlite3"

	"github.com/stregouet/nuntius/config"
	"github.com/stregouet/nuntius/database"
	_ "github.com/stregouet/nuntius/database/migrations"
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/models"
	"github.com/stregouet/nuntius/ui"
	// "github.com/stregouet/nuntius/widgets"
)

type Thread struct {
	id         int
	subject    string
	count      int
	rootmailid int
}

func (t *Thread) ToRune() []rune {
	return []rune(fmt.Sprintf("%s [%d]", t.subject, t.count))
}

func (t *Thread) FetchRoot(db *sql.DB) (*Threadline, error) {
	row := db.QueryRow("select subject, id from mail m where threadid = ? and not exists (select 1 from mail p where m.inreplyto = p.messageid)", t.id)
	var subject string
	var id int
	err := row.Scan(&subject, &id)
	if err != nil {
		return nil, err
	}
	return &Threadline{id, subject, "", 0}, nil
}

type Threadline struct {
	id      int
	subject string
	level   string
	depth   int
}

func (t *Threadline) ToRune() []rune {
	return []rune(t.subject)
}

func (t *Threadline) Depth() int {
	return t.depth
}

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

// func oldMain() {

// 	logOutput, err := os.OpenFile("/tmp/tcell.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer logOutput.Close()
// 	flog := log.New(logOutput, "logger: ", log.Lshortfile|log.Ltime)
// 	db, err := sql.Open("sqlite3", "sqlite.db")
// 	if err != nil {
// 		flog.Fatalf("cannot open db %v", err)
// 	}
// 	defer db.Close()

// 	// // XXX
// 	// // app := &views.Application{}
// 	app := ui.NewApp(flog)
// 	w := ui.NewWindow(&app, flog)
// 	app.PostRequest(&FetchMailboxes{})
// 	app.PostRequest(&FetchMailboxes{}, func(res *Response) error {
// 		res, ok := res.(MailboxesResponse)
// 		if !ok {
// 			return fmt.Errorf("error fetching mailboxes")
// 		}
// 		l := widgets.NewList(flog)
// 		for _, m := range res.Mailboxes {
// 			l.AddLine(&Mailbox{title: m})
// 		}
// 		l.OnSelect = func(line widgets.IRune) {
// 			m, _ := line.(*Mailbox)
// 			app.PostRequest(&FetchThreads{Mailbox: m.title}, func(res *Response) error {
				
// 			})
// 		}

// 	})

// 	status := views.NewSimpleStyledTextBar()
// 	status.SetStyle(tcell.StyleDefault.
// 		Background(tcell.ColorBlue).
// 		Foreground(tcell.ColorYellow))
// 	status.RegisterLeftStyle('N', tcell.StyleDefault.
// 		Background(tcell.ColorYellow).
// 		Foreground(tcell.ColorBlack))
// 	status.SetLeft("My status is here.")
// 	status.SetRight("%UCellView%N demo!")
// 	status.SetCenter("Cen%ST%Ner")

// 	l := widgets.NewList(flog)
// 	rows, err := db.Query("select threadid, max(subject), count(id), min(id) from mail group by threadid order by date desc limit 40")
// 	if err != nil {
// 		flog.Fatalf("cannot query db %v", err)
// 	}
// 	for rows.Next() {
// 		var id int
// 		var subject string
// 		var count int
// 		var mailid int
// 		rows.Scan(&id, &subject, &count, &mailid)
// 		l.AddLine(&Thread{id, subject, count, mailid})
// 	}
// 	l.OnSelect = func(line widgets.IRune) {
// 		t, _ := line.(*Thread)
// 		// t := views.NewText()
// 		// t.SetText(fmt.Sprintf("ID: %d, %s", m.id, m.subject))
// 		// w.SetContent(t)
// 		root, err := t.FetchRoot(db)
// 		if err != nil {
// 			flog.Printf("cannot found thread root %d (%v)", t.id, err)
// 			return

// 		}

// 		tree := widgets.NewTree(flog)
// 		rows, err := db.Query(`
// with recursive empdata(id, subject, messageid, level, depth) as (
//     select
//       mail.id,
//       subject,
//       messageid,
//       '/' || mail.id as level,
// 			0
//     from mail
//     where mail.id = ?
//   UNION ALL
//     select
//       this.id,
//       this.subject,
//       this.messageid,
//       prior.level || '/' || this.id,
// 			prior.depth + 1
//     from
//       empdata prior
//       inner join mail this on this.inreplyto = prior.messageid
// ) select subject, level, depth, id from empdata order by level `, root.id)
// 		if err != nil {
// 			flog.Printf("cannot select thread %d (%v)", t.id, err)
// 			return
// 		}
// 		defer rows.Close()
// 		for rows.Next() {
// 			var subject string
// 			var level string
// 			var id int
// 			var depth int
// 			rows.Scan(&subject, &level, &depth, &id)
// 			tree.AddLine(&Threadline{id, subject, level, depth})
// 		}
// 		tree.SetParent(w)
// 		w.SetContent(&tree)
// 		w.Redraw()
// 	}

// 	// w.SetStatus(status)
// 	w.SetContent(&l)

// 	app.SetStyle(tcell.StyleDefault.
// 		Foreground(tcell.ColorWhite).
// 		Background(tcell.ColorBlack))
// 	// // XXX
// 	// // app.SetRootWidget(&w)
// 	app.SetWindow(w)
// 	// if e := app.Run(); e != nil {
// 	// 	fmt.Fprintln(os.Stderr, e.Error())
// 	// }
// 	app.Run()
// 	flog.Print("exiting main")
// 	// XXX
// 	// if e := app.Run(); e != nil {
// 	// 	fmt.Fprintln(os.Stderr, e.Error())
// 	// 	os.Exit(1)
// 	// }
// }

func importdata() {
	db, err := sql.Open("sqlite3", "sqlite.db")
	if err != nil {
		log.Fatalf("cannot open db %v", err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("cannot begin tx %v", err)
	}
	err = database.Setup(tx)
	if err != nil {
		log.Fatalf("cannot setup db %v", err)
	}
	err = database.Migrate(tx)
	if err != nil {
		log.Fatalf("cannot migrate db %v", err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatalf("cannot commit %v", err)
	}

	log.Print("tables created")
	tx, err = db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Open("/home/samuel/tmp/test-golang-imap/ofbiz-dev.json")
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	i := 0
	for scanner.Scan() {
		var m models.Mail
		err = json.Unmarshal([]byte(scanner.Text()), &m)
		if err != nil {
			log.Fatal(err)
		}
		// err = m.UpdateThreadid(tx)
		// if err != nil {
		// log.Fatalf("error updating threadid `%v`", err)
		// }
		err = m.InsertInto(tx)
		if err != nil {
			log.Fatalf("error inserting mail `%v`", err)
		}
		if i%100 == 0 {
			log.Printf("indexed %d", i)
		}
		i += 1
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	tx, err = db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	roots := make([]int, 0)
	rows, err := db.Query("select m.id from mail m where not exists (select 1 from mail p where m.inreplyto = p.messageid)")
	defer rows.Close()
	if err != nil {
		log.Fatalf("cannot query db %v", err)
	}
	for rows.Next() {
		var id int
		rows.Scan(&id)
		roots = append(roots, id)
	}

	threadid := 1
	for _, id := range roots {
		m := models.Mail{Id: id}
		m.UpdateThreadidOnChild(tx, threadid)
		threadid++
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

}
