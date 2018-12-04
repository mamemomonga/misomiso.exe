package don

import (
	"log"
	"regexp"
	"sync"
	"time"
	"github.com/mattn/go-mastodon"
	"github.com/davecgh/go-spew/spew"
)

// ステータス取得チャンネル
type Gfstatus struct {
	m       string
	status  *mastodon.Status
	err     error
}

// コールバック
type LauncherCallbacks struct {
	Boost      func()
	Timeout    func(LauncherReport)
	Abort      func(LauncherReport)
}

// 報告
type LauncherReport struct {
	Hit int
	Members []string
	Remain  time.Duration
	Elapsed time.Duration
}

// 発射台設定
type LauncherConf struct {
	SearchRegexp     string
	Callbacks        LauncherCallbacks
}

// 最後のブースト情報
type LastBoosted struct {
	user    string
	content string
}

// 発射台
type Launcher struct {
	ch             chan Gfstatus
	m              *sync.Mutex
	searchRegexp   *regexp.Regexp
	cb             LauncherCallbacks
	don            *Don
	running        bool
	boost          int
	members        map[string]bool
	start_time     time.Time
	stop_time      time.Time
	wsc            *mastodon.WSClient
	jst            *time.Location
	found_ids      map[mastodon.ID]bool
	last_boosted   LastBoosted
}

func (t *Don) Launcher(lc LauncherConf)(this *Launcher, err error) {
	this     = new(Launcher)
	this.don = t
	this.cb  = lc.Callbacks

	// チャンネル
	this.ch = make(chan Gfstatus)

	// Mutex
	this.m = new(sync.Mutex)

	// 遭遇したメンバー
	this.members   = make(map[string]bool)

	// 遭遇したID
	this.found_ids = make(map[mastodon.ID]bool)

	// ブースト数
	this.boost = 0

	// 日本時間
	this.jst = time.FixedZone("Asia/Tokyo", 9*60*60)

	// 検索正規表現の設定
	log.Printf("trace: Search Regexp: %s",lc.SearchRegexp)
	this.searchRegexp = regexp.MustCompile(lc.SearchRegexp)

	// WebSocketクライアント
	this.wsc = t.client.NewWSClient()

	return this, nil
}

// Close
func (this *Launcher) Finish() {
	close(this.ch)
}

// 発射
func (this *Launcher) Run() {
	this.running = true

	// 開始時間
	this.start_time = time.Now()

	// 終了時間
	this.stop_time = this.start_time.Add( time.Second * time.Duration(this.don.config.Timeout))

	go this.chaseLTL()
	go this.chaseHTL()
	go this.detectTimeout()
}

// レシーバ
func (this *Launcher) Reciever() error {
	for {
		// チャンネル受信
		cst := <-this.ch

		// エラーを受信
		if cst.err != nil {
			return cst.err
		}

		switch cst.m {
			case "TIMEOUT":
				log.Print("info: TIMEOUT")
				this.stop()
				this.cb.Timeout( this.Report() )
				return nil

			case "ABORT":
				log.Print("info: ABORT")
				this.stop()
				this.cb.Abort( this.Report() )
				return nil
		}

		// ファボブースター
		this.favBooster(cst)
	}
}

// ファボと結果表示
func (this *Launcher) fabBoosterFabers(s *mastodon.Status) {
	// ブーストしてたら除外
	if s.Reblogged == true {
		return
	}
	// ファボってなかったらファボる
	if s.Favourited == false {
		_, err := this.don.client.Favourite(this.don.ctx, s.ID)
		if err != nil {
			log.Printf("warn: Favourite %s", err)
			spew.Dump(s)
		}
	}
	// ログ表示
	log.Printf("info: %s %s %s %s",
		s.ID,
		s.CreatedAt.In(this.jst).Format(time.RFC3339),
		s.Account.DisplayName,
		s.Content,
	)
}


// Booster
func (this *Launcher) favBooster(s Gfstatus) {

	// 自分は除外
	if s.status.Account.ID == this.don.selfId {
		return
	}

	// すでにブーストした投稿なら除外
	if this.found_ids[s.status.ID] {
		return
	}

	// ブーストしてたら除外
	if s.status.Reblogged == true {
		return
	}

	// 最後にブーストしたユーザで、同じ内容なら除外
	// if ( this.last_boosted.user == s.status.Account.URL ) && ( this.last_boosted.content == s.status.Content ) {
	// 	return
	// }

	// 最後にブーストしたユーザなら除外
	if this.last_boosted.user == s.status.Account.URL  {
		return
	}

	// 検索対象以外は除外
	if ! this.searchRegexp.MatchString(s.status.Content) {
		return
	}

	// Debug
	//log.Printf("trace: [%s] %s",s.m, spew.Sdump( s.status ))

	// ブーストされた投稿か？
	if s.status.Reblog != nil {
		this.fabBoosterFabers(s.status.Reblog)
	} else {
		this.fabBoosterFabers(s.status)
	}

	// ブーストする
	bst, err := this.don.client.Reblog(this.don.ctx, s.status.ID) // *mastodon.Status, error

	if err != nil {
		log.Printf("warn: REBLOG %s", err)
		return
	}

	this.cb.Boost()

	// ロック
	this.m.Lock()

	// 発見トゥートを記録
	this.found_ids[s.status.ID]=true

	// 最後のユーザを記録
	this.last_boosted=LastBoosted{
		user:    s.status.Account.URL,
		content: s.status.Content,
	}

	// メンバーリスト追加
	this.members[bst.Reblog.Account.DisplayName]=true

	// ブーストしたらcharge_time分期限を延長する
	this.stop_time = this.stop_time.Add( time.Second * time.Duration(this.don.config.ChargeTime))

	// カウンタをアップ
	this.boost++

	// アンロック
	this.m.Unlock()

	return
}

// 実行中
func (this *Launcher) IsRunning() bool {
	this.m.Lock()
	defer this.m.Unlock()
	return this.running
}

// 停止措置
func (this *Launcher) stop() {
	this.m.Lock()
	defer this.m.Unlock()
	this.running = false
}

// 中断指示
func (this *Launcher) Abort() {
	this.ch <- Gfstatus{m: "ABORT", status: nil, err: nil}
}

// 状況レポート
func (this *Launcher) Report()(r LauncherReport) {
	this.m.Lock()
	defer this.m.Unlock()

	n := time.Now()
	r.Hit     = this.boost
	r.Remain  = this.stop_time.Sub(n)
	r.Elapsed = n.Sub(this.start_time)
	for k,_ := range this.members {
		r.Members = append(r.Members, k)
	}

	return r
}

// LTL
func (this *Launcher) chaseLTL() {
	q, err := this.wsc.StreamingWSPublic(this.don.ctx, true)
	if err != nil {
		this.ch <- Gfstatus{m: "LTL", status: nil, err: err}
		return
	}
	for e := range q {
		if ! this.IsRunning() {
			return
		}
		if u, ok := e.(*mastodon.UpdateEvent); ok {
			this.ch <- Gfstatus{m: "LTL", status: u.Status, err: nil}
		}
	}
}

// HTL
func (this *Launcher) chaseHTL() {
	q, err := this.wsc.StreamingWSUser(this.don.ctx)
	if err != nil {
		this.ch <- Gfstatus{m: "HTL", status: nil, err: err}
		return
	}
	for e := range q {
		if ! this.IsRunning() {
			return
		}
		if u, ok := e.(*mastodon.UpdateEvent); ok {
			this.ch <- Gfstatus{m: "HTL", status: u.Status, err: nil}
		}
	}

}

// Timeout
func (this *Launcher) detectTimeout() {
	for {
		if ! this.IsRunning() {
			return
		}

		this.m.Lock()
		r := this.stop_time.Sub(time.Now())
		this.m.Unlock()

		// debug
		//log.Printf("trace: Remain %3.2f sec", r.Seconds())

		if r <= 0 {
			this.ch <- Gfstatus{m: "TIMEOUT", status: nil, err: nil}
		}

		time.Sleep(time.Millisecond * 100)
	}
}


