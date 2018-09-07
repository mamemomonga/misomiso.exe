package main

import (
	"./config"
	"./don"
	"./store"
	"fmt"
	"github.com/comail/colog"
	"log"
	"os"
	//	"github.com/davecgh/go-spew/spew"
)

var (
	Version  string
	Revision string
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {

	// colog 設定
	if Version == "" {
		colog.SetDefaultLevel(colog.LDebug)
		colog.SetMinLevel(colog.LTrace)
		colog.SetFormatter(&colog.StdFormatter{
			Colors: true,
			Flag:   log.Ldate | log.Ltime | log.Lshortfile,
		})
	} else {
		colog.SetDefaultLevel(colog.LDebug)
		colog.SetMinLevel(colog.LInfo)
		colog.SetFormatter(&colog.StdFormatter{
			Colors: true,
			Flag:   log.Ldate | log.Ltime | log.Lshortfile,
		})
	}
	colog.Register()

	// スタートアップ
	log.Printf("info: misomiso.exe version %s", Version)

	if len(args) == 0 {
		usage()
		return 1
	}

	var err error

	// config 読込
	var cnf config.Config
	cnf, err = config.Load(os.Args[1])
	if err != nil {
		log.Printf("alert: %s", err)
		return 1
	}

	// data読書
	var stor *store.Store
	stor, err = store.NewStore(cnf.DataFile)
	if err != nil {
		log.Print("alert: %s", err)
		return 1
	}

	// マストドン
	var dn *don.Don
	dn, err = don.NewDon(&cnf, stor)
	if err != nil {
		log.Print("alert: %s", err)
		return 1
	}

	// トゥート
	err = dn.Toot("みそみそ〜")
	if err != nil {
		log.Print("alert: %s", err)
		return 1
	}

	return 0
}

func usage() {
	fmt.Printf("\n\n  USAGE: %s config.yaml\n\n", os.Args[0])
}
