package options

import (
	"encoding/base64"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"time"
)

// HopeConfig настройки приложения
var HopeConfig = struct {
	Core    *ConfigCore     `yaml:"core"`
	Host    *ConfigHost     `yaml:"host"`
	Scripts []*ConfigScript `yaml:"scripts"`
}{
	Core: &ConfigCore{
		Workers:  13,
		Requests: 8000,
		StatUrl:  "http://127.0.0.1/usage/statistic.json",
	},
}

// ConfigCore базовые настройки теста
type ConfigCore struct {
	Workers  int           `yaml:"workers"`
	Requests int           `yaml:"connections"`
	Duration time.Duration `yaml:"duration"`
	StatUrl  string        `yaml:"stat_url"`
}

// ConfigHost описание обстреливаемого сервера
type ConfigHost struct {
	Address  string      `yaml:"addr"`
	Port     int         `yaml:"port"`
	Protocol string      `yaml:"protocol"`
	Header   http.Header `yaml:"header"`
}

// ConfigScript сценарий для нагрузки
type ConfigScript struct {
	Name     string          `yaml:"name"`
	Header   http.Header     `yaml:"header,omitempty"`
	Requests []ConfigRequest `yaml:"requests"`
}

// ConfigRequest Запрс исполняемый на севере
type ConfigRequest struct {
	Resource string      `yaml:"resource"`
	Method   string      `yaml:"method,omitempty"`
	Body     []byte      `yaml:"body,omitempty"`
	Rate     int         `yamle:"rate,omitempty"`
	Header   http.Header `yaml:"header,omitempty"`
}

// Реализация десириализации коры
func (c *ConfigCore) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tc struct {
		Workers  int    `yaml:"workers"`
		Requests int    `yaml:"requests"`
		Duration string `yaml:"duration"`
	}
	var e error

	if e = unmarshal(&tc); e != nil {
		return e
	}

	if c.Duration, e = time.ParseDuration(tc.Duration); e != nil {
		return e
	}
	c.Workers = tc.Workers
	c.Requests = tc.Requests

	return nil
}

type headerEntry struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
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
		Port     int           `yaml:"port,omitempty"`
		Protocol string        `yaml:"protocol,omitempty"`
		Header   []headerEntry `yaml:"header,omitempty"`
	}

	tc.Port = 80
	tc.Protocol = "HTTP"

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

func (c *ConfigRequest) UnmarshalYAML(unmarshal func(interface{}) error) (e error) {
	var tc struct {
		Resource string        `yaml:"resource"`
		Method   string        `yaml:"method,omitempty"`
		Body     string        `yaml:"body,omitempty"`
		Rate     int           `yamle:"rate,omitempty"`
		Header   []headerEntry `yaml:"header,omitempty"`
	}
	e = nil
	tc.Method = "GET"
	tc.Rate = 1

	if e = unmarshal(&tc); e != nil {
		return
	}

	c.Resource = tc.Resource
	c.Method = tc.Method
	if tc.Body != "" {
		if c.Body, e = base64.StdEncoding.DecodeString(tc.Body); e != nil {
			return
		}
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
		for k := range header {
			h.Set(k, header.Get(k))
		}
	}
	return h
}

// Непосредственно про чтение конфига

// LoadConfig глобальные настройки конфигурации

func LoadConfigFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &HopeConfig)
}
