package main

import (
	"./config"
	"./don"
	"./store"
	// "fmt"
	"github.com/comail/colog"
	"log"
	"os"
	"path/filepath"
	//	"github.com/davecgh/go-spew/spew"
	"flag"
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
			// Flag:   log.Ldate | log.Ltime | log.Lshortfile,
			Flag:   log.Ldate | log.Ltime,
		})
	}
	colog.Register()

	appdir := filepath.Dir(os.Args[0])

	// スタートアップ
	log.Printf("info: misomiso.exe version %s", Version)

	var err error

	// config 読込
	var cnf config.Config
	cnf, err = config.Load(filepath.Join(appdir,"config.yaml"))
	if err != nil {
		log.Printf("alert: %s", err)
		return 1
	}

	// フラグの取得
	{
		// -t 開始宣言トゥート
		var t = flag.String("t", "みそみそ〜", "Start Up toot")

		// -r 検索正規表現
		var r = flag.String("r", "みそみそ", "Search Regexp")

		flag.Parse()
		cnf.StartupToot = *t
		cnf.SearchRegexp = *r
	}

	log.Printf("info: STARTUP %s",cnf.StartupToot)
	log.Printf("info: REGEXP  %s",cnf.SearchRegexp)

	// data読書
	var stor *store.Store
	stor, err = store.NewStore(filepath.Join(appdir,"data.yaml"))
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

	// みそリスナー
	err = dn.MisoListener()
	if err != nil {
		log.Print("alert: %s", err)
		return 1
	}

	return 0
}

