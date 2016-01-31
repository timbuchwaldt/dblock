package dblockmain

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

type Config struct {
	Files     map[string]LogFileConfig
	Whitelist []string
}

type LogFileConfig struct {
	Filename string
	Regexes  []string
}

func ParseConfig(file string) Config {
	tomlData, err := ioutil.ReadFile(file)
	check(err)
	var conf Config
	toml.Decode(string(tomlData[:]), &conf)
	return conf
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
