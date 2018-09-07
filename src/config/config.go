package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	DataFile   string    `yaml:"data_file"`
	ClientName string    `yaml:"client_name"`
	Mastodon   CMastodon `yaml:"mastodon"`
	Message    string    `yaml:"message"`
}

type CMastodon struct {
	Domain   string
	Email    string
	Password string
}

func Load(filename string) (cnf Config, err error) {

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return cnf, err
	}

	err = yaml.Unmarshal(buf, &cnf)
	if err != nil {
		return cnf, err
	}

	log.Printf("trace: Load %s", filename)

	return cnf, nil
}
