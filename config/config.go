package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"time"
)

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

type Config struct {
	Files         map[string]LogFileConfig
	Whitelist     []string
	IncidentTime  duration
	MaxIncidents  int
	EtcdAddresses []string
}

type LogFileConfig struct {
	Filename string
	Regexes  []string
}

func ParseConfig(file string) Config {
	tomlData, err := ioutil.ReadFile(file)
	check(err)
	var conf Config
	_, err = toml.Decode(string(tomlData[:]), &conf)
	check(err)
	return conf
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
