package don

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	DataFile     string    `yaml:"data_file"`
	ClientName   string    `yaml:"client_name"`
	Timeout      int       `yaml:"timeout"`
	Mastodon     CMastodon `yaml:"mastodon"`
}

type CMastodon struct {
	Domain   string
	Email    string
	Password string
}

func (this *Don) loadConfig(filename string) error {

	var cnf Config

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, &cnf)
	if err != nil {
		return err
	}

	log.Printf("trace: Load %s", filename)
	this.config = cnf
	return nil
}
