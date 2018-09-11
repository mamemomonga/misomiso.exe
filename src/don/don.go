package don

import (
	"../config"
	"../store"
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/mattn/go-mastodon"
	"log"
	"regexp"
	"sync"
	"time"
)

type Don struct {
	ctx           context.Context
	client        *mastodon.Client
	client_id     string
	client_secret string
	jst           *time.Location
	startup_toot  string
	search_regexp string
	timeout       timeoutCounter
}

type timeoutCounter struct {
	t time.Time
	m *sync.Mutex
}

func NewDon(cnf *config.Config, stor *store.Store) (this *Don, err error) {
	this = new(Don)
	this.ctx = context.Background()
	this.client_id = ""
	this.client_secret = ""

	this.startup_toot = cnf.StartupToot
	this.search_regexp = cnf.SearchRegexp

	domain := cnf.Mastodon.Domain
	email := cnf.Mastodon.Email
	password := cnf.Mastodon.Password
	client_name := "misomiso.exe"

	if val, ok := stor.Data.Apps[cnf.Mastodon.Domain]; ok {
		this.client_id = val.ClientID
		this.client_secret = val.ClientSecret
	} else {
		// アプリケーション登録
		app, err := mastodon.RegisterApp(this.ctx, &mastodon.AppConfig{
			Server:     fmt.Sprintf("https://%s/", domain),
			ClientName: client_name,
			Scopes:     "read write follow",
		})
		if err != nil {
			return this, err
		}
		this.client_id = app.ClientID
		this.client_secret = app.ClientSecret
		stor.Data.Apps[domain] = store.App{
			ClientID:     app.ClientID,
			ClientSecret: app.ClientSecret,
		}
		stor.Save()
		log.Printf("info: Register App %s", domain)
	}

	// クライアント
	this.client = mastodon.NewClient(&mastodon.Config{
		Server:       fmt.Sprintf("https://%s/", domain),
		ClientID:     this.client_id,
		ClientSecret: this.client_secret,
	})
	err = this.client.Authenticate(this.ctx, email, password)
	if err != nil {
		return this, err
	}

	// 日本時間
	this.jst = time.FixedZone("Asia/Tokyo", 9*60*60)

	log.Printf("info: Start Clients %s", domain)
	return
}

// みそリスナー
func (this *Don) MisoListener() (err error) {

	// 自分のIDを得る
	var self_id mastodon.ID
	{
		var account *mastodon.Account
		account, err = this.client.GetAccountCurrentUser(this.ctx)
		self_id = account.ID
	}
	log.Printf("SelfID: %s", self_id)

	// みそみそを宣言する
	{
		toot := mastodon.Toot{Status: fmt.Sprintf("%s #misomiso", this.startup_toot)}
		_, err := this.client.PostStatus(this.ctx, &toot)
		if err != nil {
			return err
		}
		log.Printf("info: Toot %s", toot.Status)
	}

	// 検索正規表現の設定
	r := regexp.MustCompile(this.search_regexp)

	// ステータス取得チャンネル
	type chStatus struct {
		m      string
		status *mastodon.Status
		err    error
	}
	chs := make(chan chStatus)
	defer close(chs)

	// WebSocketクライアント
	wsc := this.client.NewWSClient()

	// LTL接続
	go func() {
		q, err := wsc.StreamingWSPublic(this.ctx, true)
		if err != nil {
			chs <- chStatus{m: "LTL", status: nil, err: err}
			return
		}
		for e := range q {
			if u, ok := e.(*mastodon.UpdateEvent); ok {
				chs <- chStatus{m: "LTL", status: u.Status, err: nil}
			}
		}
		return
	}()

	// HTL接続
	go func() {
		q, err := wsc.StreamingWSUser(this.ctx)
		if err != nil {
			chs <- chStatus{m: "HTL", status: nil, err: err}
			return
		}
		for e := range q {
			if u, ok := e.(*mastodon.UpdateEvent); ok {
				chs <- chStatus{m: "HTL", status: u.Status, err: nil}
			}
		}
		return
	}()

	this.timeout = timeoutCounter{
		t: time.Now(),
		m: new(sync.Mutex),
	}

	// タイムアウト
	go func() {
		for {
			this.timeout.m.Lock()
			dur := time.Now().Sub(this.timeout.t)
			this.timeout.m.Unlock()

			log.Printf("trace: Wait %3.0f sec", dur.Seconds())
			// 1分間トゥートがみつからなかったら終了
			if dur > (time.Second * 60 * 1) {
				chs <- chStatus{m: "TIMEOUT", status: nil, err: nil}
			}
			time.Sleep(time.Second)
		}
	}()

	for {
		cst := <-chs

		// エラーを受信
		if cst.err != nil {
			return cst.err
		}

		// タイムアウトを受信
		if cst.m == "TIMEOUT" {
			log.Print("info: TIMEOUT")
			return
		}

		// log.Printf(" --- timeline: %s ---",cst.timeline)
		// log.Print(cst.status.Content)

		// みそ喰い
		err = this.miso_eater(cst.status, self_id, r)
		if err != nil {
			return
		}
	}
	return nil
}

// みそ喰い
func (this *Don) miso_eater(status *mastodon.Status, self_id mastodon.ID, r *regexp.Regexp) (err error) {

	// ファボと結果表示
	fabres := func(s *mastodon.Status) {

		// ブーストしてたら除外
		if s.Reblogged == true {
			return
		}

		// ファボってなかったらファボる
		if s.Favourited == false {
			_, err := this.client.Favourite(this.ctx, s.ID)
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
	if status.Account.ID == self_id {
		return nil
	}

	// 検索対象以外は除外
	if !r.MatchString(status.Content) {
		return nil
	}

	// ブーストしてたら除外
	if status.Reblogged == true {
		return nil
	}

	// ブーストされた投稿か？
	if status.Reblog != nil {
		fabres(status.Reblog)
	} else {
		fabres(status)
	}

	// ブーストする
	_, err = this.client.Reblog(this.ctx, status.ID)
	this.timeout.m.Lock()
	this.timeout.t = time.Now()
	this.timeout.m.Unlock()
	if err != nil {
		log.Printf("warn: Reblog %s", err)
	}

	return nil
}
