package main

// みそパンチ！

import (
	"log"
	"time"
	"fmt"
	"strings"
	"github.com/mamemomonga/misomiso.exe/go/don"
//	"github.com/davecgh/go-spew/spew"
)

// レポート時間(秒)
const ReportInterval = 15

type MisoPunch struct {
	d       *don.Don
	l       *don.Launcher
	Keyword string
	Regexp  string
}

func NewMisoPunch(config_file string, keyword string, regexp string)(this *MisoPunch,err error) {
	this = new(MisoPunch)
	this.Keyword = keyword
	this.Regexp  = regexp

	d, err := don.NewDon(config_file)
	if err != nil {
		return this,err
	}
	this.d = d

	err = this.d.Connect()
	if err != nil {
		return this,err
	}
	return this, nil
}

func (this *MisoPunch) Run() error {
	l, err := this.d.Launcher( don.LauncherConf{
		SearchRegexp: this.Regexp,
		Callbacks : don.LauncherCallbacks{
			Boost:   this.cBoosted,
			Timeout: this.cReportFinish,
			Abort:   this.cReportFinish,
		},
	})
	if err != nil {
		return err
	}
	this.l = l

	err = this.d.Toot(fmt.Sprintf("[発射]\n%s #misomiso", this.Keyword))
	if err != nil {
		return err
	}

	this.l.Run()

	go func() {
		for {
			if ! this.l.IsRunning() {
				break
			}
			time.Sleep(time.Second * ReportInterval)
			this.cReportRunning( this.l.Report() )
		}
	}()

	err = this.l.Reciever()
	if err != nil {
		return err
	}

	return nil
}

func (this *MisoPunch) Finish() {
	this.l.Finish()
}

func (this *MisoPunch) cBoosted() {
	log.Print("info: ブーストしました")
//	this.d.Toot(fmt.Sprintf("%s を捕捉しました #misomiso", this.Keyword))
}

func (this *MisoPunch) cReportFinish(r don.LauncherReport) {

	mbs := []string{}
	for _,i := range r.Members {
		mbs = append(mbs, i+" さん")
	}
	mb := strings.Join( mbs,", ")

	e := fmt.Sprintf("%.0f分%02d秒",r.Elapsed.Truncate(time.Minute).Minutes(), int(r.Elapsed.Seconds()) % 60 )

	var n string
	if r.Hit == 0 {
		n = fmt.Sprintf("[自爆]\n戦果: %d%s\n飛行時間: %s\n追尾終了しました。#misomiso",
		r.Hit, this.Keyword, e )

	} else {
		n = fmt.Sprintf("[自爆]\n戦果: %d%s\n飛行時間: %s\n追尾終了しました。\n%s ありがとうございました。#misomiso",
			r.Hit, this.Keyword, e, mb )
	}

	this.d.Toot(n)
}

func (this *MisoPunch) cReportRunning(r don.LauncherReport) {
	m := fmt.Sprintf("%.0f分%02d秒",r.Remain.Truncate(time.Minute).Minutes(),  int(r.Remain.Seconds()) % 60 )
	e := fmt.Sprintf("%.0f分%02d秒",r.Elapsed.Truncate(time.Minute).Minutes(), int(r.Elapsed.Seconds()) % 60 )
	n := fmt.Sprintf("[追尾中]\n戦果: %d%s\n残り時間: %s\n経過時間: %s\n#misomiso", r.Hit, this.Keyword, m, e)
	this.d.Toot(n)
}

