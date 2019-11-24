package options

import (
	"encoding/base64"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"time"
)

// ConfigCore базовые настройки теста
type ConfigCore struct {
	Thread      int           `yaml:"threads"`
	Connections int           `yaml:"threads"`
	Duration    time.Duration `yaml:"duration"`
}

// Реализация десириализации коры
func (c *ConfigCore) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tc struct {
		Thread      int    `yaml:"threads"`
		Connections int    `yaml:"connections"`
		Duration    string `yaml:"duration"`
	}
	var e error

	if e = unmarshal(&tc); e != nil {
		return e
	}

	if c.Duration, e = time.ParseDuration(tc.Duration); e != nil {
		return e
	}
	c.Thread = tc.Thread
	c.Connections = tc.Connections

	return nil
}

type headerEntry struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type ConfigHost struct {
	Address  string      `yaml:"addr"`
	Port     int         `yaml:"port"`
	Protocol string      `yaml:"protocol"`
	Header   http.Header `yaml:"header"`
}

func (c *ConfigHost) GetUrlFor(r *ConfigRequest) string {
	ssl := false
	url := "http://"
	if c.Protocol == "https" || c.Protocol == "https2" {
		ssl = true
		url = "https://"
	}
	if (!ssl && c.Port == 80) || (ssl && c.Port == 443) {
		url += c.Address
	} else {
		url += fmt.Sprintf("%s:%d", c.Address, c.Port)
	}
	url += r.Resource
	return url
}

func (c *ConfigHost) UnmarshalYaml(unmarshal func(interface{}) error) error {
	var tc struct {
		Address  string        `yaml:"addr"`
		Port     int           `yaml:"port"`
		Protocol string        `yaml:"protocol"`
		Header   []headerEntry `yaml:"header,omitempty"`
	}

	if e := unmarshal(&tc); e != nil {
		return e
	}
	c.Address = tc.Address
	c.Port = tc.Port
	c.Protocol = tc.Protocol

	c.Header = make(http.Header)
	for _, ent := range tc.Header {
		c.Header.Add(ent.Name, ent.Value)
	}
	return nil
}

type ConfigScript struct {
	Name     string          `yaml:"name"`
	Header   http.Header     `yaml:"header,omitempty"`
	Requests []ConfigRequest `yaml:"requests"`
}

func (c *ConfigScript) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tc struct {
		Name     string          `yaml:"name"`
		Header   []headerEntry   `yaml:"header,omitempty"`
		Requests []ConfigRequest `yaml:"requests"`
	}

	if e := unmarshal(&tc); e != nil {
		return e
	}
	c.Name = tc.Name
	c.Requests = tc.Requests
	c.Header = make(http.Header)
	for _, ent := range tc.Header {
		c.Header.Add(ent.Name, ent.Value)
	}
	return nil
}

type ConfigRequest struct {
	Resource string      `yaml:"resource"`
	Method   string      `yaml:"method,omitempty"`
	Body     []byte      `yaml:"body,omitempty"`
	Rate     int         `yamle:"rate,omitempty"`
	Header   http.Header `yaml:"header,omitempty"`
}

func (c *ConfigRequest) UnmarshalYAML(unmarshal func(interface{}) error) (e error) {
	var tc struct {
		Resource string        `yaml:"resource"`
		Method   string        `yaml:"method,omitempty"`
		Body     string        `yaml:"body,omitempty"`
		Rate     int           `yamle:"rate,omitempty"`
		Header   []headerEntry `yaml:"header,omitempty"`
	}
	e = nil

	if e = unmarshal(&tc); e != nil {
		return
	}

	c.Resource = tc.Resource
	c.Method = tc.Method
	if c.Body, e = base64.StdEncoding.DecodeString(tc.Body); e != nil {
		return
	}
	c.Rate = tc.Rate
	c.Header = make(http.Header)
	for _, ent := range tc.Header {
		if c.Header == nil {
			panic("Nil header WAT!")
		}
		c.Header.Set(ent.Name, ent.Value)
	}
	return
}

func MergeHttpHeaders(headers ...http.Header) http.Header {
	h := http.Header{}

	for _, header := range headers {
		for k, _ := range header {
			h.Set(k, header.Get(k))
		}
	}
	return h
}

// Непосредственно про чтение конфига

// LoadConfig глобальные настройки конфигурации
var HopeConfig struct {
	Core    *ConfigCore     `yaml:"core"`
	Host    *ConfigHost     `yaml:"host"`
	Scripts []*ConfigScript `yaml:"scripts"`
}

func LoadConfigFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &HopeConfig)
}
