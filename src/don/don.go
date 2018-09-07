package don

import (
	"../config"
	"../store"
	"context"
	"fmt"
	"github.com/mattn/go-mastodon"
	"log"
	//	"github.com/davecgh/go-spew/spew"
)

type Don struct {
	ctx           context.Context
	client        *mastodon.Client
	client_id     string
	client_secret string
}

type Config struct {
}

func NewDon(cnf *config.Config, stor *store.Store) (this *Don, err error) {
	this = new(Don)
	this.ctx = context.Background()
	this.client_id = ""
	this.client_secret = ""

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
	log.Printf("info: Mastodon Client Start %s", domain)
	return
}

// トゥート
func (this *Don) Toot(message string) (err error) {
	toot := mastodon.Toot{Status: message}
	status, err := this.client.PostStatus(this.ctx, &toot)
	if err != nil {
		return err
	}

	_ = status
	// spew.Dump(status)
	log.Printf("info: Toot %s", message)

	return nil
}
