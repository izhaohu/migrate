package main

import (
	"flag"
	"os"

	"github.com/libgo/logx"

	"github.com/izhaohu/migrate/database"
	_ "github.com/izhaohu/migrate/database/mysql"
	"github.com/izhaohu/migrate/internal"
	"github.com/izhaohu/migrate/source"
	_ "github.com/izhaohu/migrate/source/file"
)

var (
	d  string
	p  string
	m  string
	dg bool
)

func init() {
	// Register dsn format -> [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	flag.StringVar(&d, "d", "root:PassWord@tcp(192.168.0.1:3306)/dolphin?multiStatements=true", "database uri")
	// Path to migrate files
	flag.StringVar(&p, "p", "./migrate", "migration source file path")
	// Module to up, default is "all"
	flag.StringVar(&m, "m", "all", "module to up")
	// Debug flag
	flag.BoolVar(&dg, "D", false, "dubug")
}

func main() {
	flag.Parse()

	logx.SetGlobalLevel(logx.InfoLevel)
	if dg {
		logx.SetGlobalLevel(logx.DebugLevel)
	}

	db, err := database.Open("mysql", d)
	if err != nil {
		logx.Errorf("migrate: open database error: %s", err.Error())
		os.Exit(1)
	}

	r, err := source.Open("file", p)
	if err != nil {
		logx.Errorf("migrate: read source migration error: %s", err.Error())
		os.Exit(1)
	}

	mig := internal.New(r, db)
	if m == "all" {
		err = mig.UpAll()
	} else {
		err = mig.Up(source.Module(m))
	}

	if err != nil {
		logx.Errorf("migrate: migrating error: %s", err.Error())
		os.Exit(1)
	}
}
