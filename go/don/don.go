package don

import (
	"github.com/mattn/go-mastodon"
	"log"
	"context"
	"fmt"
	// "github.com/davecgh/go-spew/spew"
)

type Don struct {
	config   Config
	store    *Store

	ctx           context.Context
	client        *mastodon.Client
	client_id     string
	client_secret string

	selfId mastodon.ID
}

// 新規作成
func NewDon(config_file string)(this *Don, err error) {
	this = new(Don)

	// Config
	err = this.loadConfig(config_file)
	if err != nil {
		return this, err
	}

	// config.yaml の場合同じ場所に config_store.yaml を生成
	store_file := filename_add_suffix(config_file,"_store")

	// Store
	stor, err := NewStore(store_file)
	if err != nil {
		return this, err
	}
	this.store = stor

	this.ctx = context.Background()
	this.client_id = ""
	this.client_secret = ""

	return this, nil
}

// 接続
func (this *Don) Connect() error {

	domain      := this.config.Mastodon.Domain
	email       := this.config.Mastodon.Email
	password    := this.config.Mastodon.Password

	if val, ok := this.store.Data.Apps[this.config.Mastodon.Domain]; ok {
		// アプリを登録済み
		this.client_id     = val.ClientID
		this.client_secret = val.ClientSecret
	} else {
		// アプリ登録
		app, err := mastodon.RegisterApp(this.ctx, &mastodon.AppConfig{
			Server:     fmt.Sprintf("https://%s/", domain),
			ClientName: this.config.ClientName,
			Scopes:     "read write follow",
		})
		if err != nil {
			return err
		}
		this.client_id     = app.ClientID
		this.client_secret = app.ClientSecret
		this.store.Data.Apps[domain] = SApp{
			ClientID:     app.ClientID,
			ClientSecret: app.ClientSecret,
		}
		this.store.Save()
		log.Printf("info: Register App %s", domain)
	}

	// クライアント
	this.client = mastodon.NewClient(&mastodon.Config{
		Server:       fmt.Sprintf("https://%s/", domain),
		ClientID:     this.client_id,
		ClientSecret: this.client_secret,
	})

	// 認証
	err := this.client.Authenticate(this.ctx, email, password)
	if err != nil {
		return err
	}

	// 自分のIDを得る
	{
		var account *mastodon.Account
		account, err = this.client.GetAccountCurrentUser(this.ctx)

		log.Printf("info: Domain:      %s",domain)
		log.Printf("info: SelfID:      %s",account.ID)
		log.Printf("info: Username:    %s",account.Username)
		log.Printf("info: DisplayName: %s",account.DisplayName)

		this.selfId = account.ID
	}

	return nil
}

// トゥートする
func (this *Don) Toot(s string) error {
	toot := mastodon.Toot{ Status: s }
	_, err := this.client.PostStatus( this.ctx, &toot )
	if err != nil {
		return err
	}
	log.Printf("info: Toot %s", toot.Status)
	return nil
}

