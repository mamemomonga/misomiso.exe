package main

import (
	"log"
	"flag"
)

func run() error {
	log.Printf("info: misomiso.exe VERSION:%s REVISION:%s",Version,Revision)

	// 設定ファイル
	config_file := flag.String("config","etc/config.yaml","Config File")

	// ターゲット
	target := flag.String("target", "みそみそ〜", "Start Up toot")

	// 正規表現
	regexp := flag.String("regexp", "みそみそ", "Search Regexp")

	flag.Parse()

	m,err := NewMisoPunch( *config_file, *target, *regexp )
	if err != nil {
		return err
	}
	defer m.Finish()

	err = m.Run()
	if err != nil {
		return err
	}

	return nil
}
