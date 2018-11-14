package don

import (
	"log"
	"regexp"
	"sync"
	"time"
	"github.com/mattn/go-mastodon"
	"github.com/davecgh/go-spew/spew"
)

// タイムアウトカウンタ
type timeoutCounter struct {
	t time.Time
	m *sync.Mutex
}

// ブーストカウンタ
type boostCounter struct {
	c int
	m *sync.Mutex
}

// ステータス取得チャンネル
type chStatus struct {
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

// 発射台
type Launcher struct {
	searchRegexp *regexp.Regexp
	timeOutTime  time.Duration

	cb         LauncherCallbacks
	running    bool
	running_m  *sync.Mutex
	don        *Don
	chs        chan chStatus
	boost      boostCounter
	timeout    timeoutCounter
	wsc        *mastodon.WSClient
	jst        *time.Location
	start_time time.Time

	members    map[string]bool
	members_m  *sync.Mutex
}

func (t *Don) Launcher(lc LauncherConf)(this *Launcher, err error) {
	this     = new(Launcher)
	this.don = t
	this.cb  = lc.Callbacks
	this.timeOutTime = time.Second * time.Duration( this.don.config.Timeout )
	this.running_m = new(sync.Mutex)

	this.members = make(map[string]bool)
	this.members_m = new(sync.Mutex)

	// 日本時間
	this.jst = time.FixedZone("Asia/Tokyo", 9*60*60)

	// 検索正規表現の設定
	log.Printf("trace: Search Regexp: %s",lc.SearchRegexp)
	this.searchRegexp = regexp.MustCompile(lc.SearchRegexp)

	// ステータス取得チャンネル
	this.chs = make(chan chStatus)

	// タイムアウトカウンタ
	this.timeout = timeoutCounter{
		t: time.Now(),
		m: new(sync.Mutex),
	}

	// ブーストカウンター
	this.boost = boostCounter{
		c: 0,
		m: new(sync.Mutex),
	}

	// WebSocketクライアント
	this.wsc = t.client.NewWSClient()

	return this, nil
}

// 発射
func (this *Launcher) Run() {
	this.running = true
	this.start_time = time.Now()
	go this.chaseLTL()
	go this.chaseHTL()
	go this.detectTimeout()
}

// レシーバ
func (this *Launcher) Reciever() error {
	for {
		// チャンネル受信
		cst := <-this.chs

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
		this.favBooster(cst.status)
	}
}

// 停止措置
func (this *Launcher) stop() {
	this.running_m.Lock()
	this.running = false
	this.running_m.Unlock()
}

// 中断指示
func (this *Launcher) Abort() {
	this.chs <- chStatus{m: "ABORT", status: nil, err: nil}
}

// 実行中
func (this *Launcher) IsRunning() bool {
	r := false
	this.running_m.Lock()
	r = this.running
	this.running_m.Unlock()
	return r
}

// 状況レポート
func (this *Launcher) Report()(r LauncherReport) {

	this.boost.m.Lock()
	r.Hit = this.boost.c
	this.boost.m.Unlock()

	this.timeout.m.Lock()
	r.Remain =  this.timeOutTime - time.Now().Sub(this.timeout.t)
	this.timeout.m.Unlock()

	this.members_m.Lock()
	for k,_ := range this.members {
		r.Members = append(r.Members, k)
	}
	this.members_m.Unlock()

	r.Elapsed = time.Now().Sub(this.start_time)
	return r
}

// LTL
func (this *Launcher) chaseLTL() {
	q, err := this.wsc.StreamingWSPublic(this.don.ctx, true)
	if err != nil {
		this.chs <- chStatus{m: "LTL", status: nil, err: err}
		return
	}
	for e := range q {
		if ! this.IsRunning() {
			return
		}
		if u, ok := e.(*mastodon.UpdateEvent); ok {
			this.chs <- chStatus{m: "LTL", status: u.Status, err: nil}
		}
	}
}

// HTL
func (this *Launcher) chaseHTL() {
	q, err := this.wsc.StreamingWSUser(this.don.ctx)
	if err != nil {
		this.chs <- chStatus{m: "HTL", status: nil, err: err}
		return
	}
	for e := range q {
		if ! this.IsRunning() {
			return
		}
		if u, ok := e.(*mastodon.UpdateEvent); ok {
			this.chs <- chStatus{m: "HTL", status: u.Status, err: nil}
		}
	}

}

// Timeout
func (this *Launcher) detectTimeout() {
	for {
		if ! this.IsRunning() {
			return
		}

		this.timeout.m.Lock()
		dur := time.Now().Sub(this.timeout.t)
		this.timeout.m.Unlock()

		// log.Printf("trace: Wait %3.2f sec", dur.Seconds())

		if dur > this.timeOutTime {
			this.chs <- chStatus{m: "TIMEOUT", status: nil, err: nil}
		}

		time.Sleep(time.Millisecond * 100)
	}
}

// Booster
func (this *Launcher) favBooster(status *mastodon.Status) {

	// ファボと結果表示
	fabres := func(s *mastodon.Status) {

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

	// 自分は除外
	if status.Account.ID == this.don.selfId {
		return
	}

	// 検索対象以外は除外
	if ! this.searchRegexp.MatchString(status.Content) {
		return
	}

	// ブーストしてたら除外
	if status.Reblogged == true {
		return
	}

	// ブーストされた投稿か？
	if status.Reblog != nil {
		fabres(status.Reblog)
	} else {
		fabres(status)
	}

	// ブーストする
	bst, err := this.don.client.Reblog(this.don.ctx, status.ID) // *mastodon.Status, error
	if err != nil {
		log.Printf("warn: Reblog %s", err)
	}
	// spew.Dump(bst)

	this.cb.Boost()

	this.members_m.Lock()
	this.members[bst.Reblog.Account.DisplayName]=true
	this.members_m.Unlock()


	// ブーストしたのでタイムアウトをリセット
	this.timeout.m.Lock()
	this.timeout.t = time.Now()
	this.timeout.m.Unlock()

	// カウンタをアップ
	this.boost.m.Lock()
	this.boost.c++
	this.boost.m.Unlock()


	return
}

// Close
func (this *Launcher) Finish() {
	close(this.chs)
}

