package store

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Store struct {
	filename string
	Data     Data
}

type Data struct {
	Apps map[string]App
}

type App struct {
	ClientID     string
	ClientSecret string
}

func NewStore(filename string) (this *Store, err error) {
	err = nil
	this = new(Store)

	this.filename = filename
	this.Data = Data{map[string]App{}}

	if b, _ := exists(filename); b == false {
		log.Printf("trace: Create %s", filename)
		this.Save()
	}

	err = this.read(filename, &this.Data)
	if err != nil {
		return
	}

	log.Printf("trace: Load %s", filename)
	return
}

func (this *Store) Save() (err error) {
	err = this.write(this.filename, this.Data)
	log.Printf("trace: Save %s", this.filename)
	return
}

func (this *Store) read(filename string, cnf interface{}) (err error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(buf, cnf)
	if err != nil {
		return
	}
	return nil
}

func (this *Store) write(filename string, cnf interface{}) (err error) {
	buf, err := yaml.Marshal(&cnf)
	if err != nil {
		return
	}
	ioutil.WriteFile(filename, buf, 0644)
	return
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
