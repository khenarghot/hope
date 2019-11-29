package options

import (
	"bytes"
	yaml "gopkg.in/yaml.v2"
	"net/http"
	"testing"
)

var headerConfig = []byte(`
name: Content-Type
value: text/json
`)

func TestHeaderEntry(t *testing.T) {
	cfg := headerEntry{}

	if err := yaml.Unmarshal(headerConfig, &cfg); err != nil {
		t.Errorf("Can't unmarsha cfg: %s", err.Error())
	}
	if cfg.Name != "Content-Type" {
		t.Errorf("Wrong Name value: '%s' instead Content-Type", cfg.Name)
	}
	if cfg.Value != "text/json" {
		t.Errorf("Wrong Value value: '%s' instead text/json", cfg.Value)
	}
}

var requestConfig = []byte(`
resource: /hls/CH_UFCHD_HLS/master.m3u8
method: POST
body:   dXNlcjpwYXNzd29yZA==
rate:   42
header:
    - name: X-TYPE
      value: Test
    - name: X-SIZE
      value: empty
`)

func TestConfigRequest(t *testing.T) {
	cfg := ConfigRequest{}

	if err := yaml.Unmarshal(requestConfig, &cfg); err != nil {
		t.Errorf("Can't unmarsha cfg: %s", err.Error())
	}

	if cfg.Resource != "/hls/CH_UFCHD_HLS/master.m3u8" {
		t.Errorf("Wrong resource '%s'", cfg.Resource)
	}
	if cfg.Method != "POST" {
		t.Errorf("Wrong method '%s'", cfg.Method)
	}
	if bytes.Equal(cfg.Body, []byte(`user:password\0`)) {
		t.Errorf("Wrong body %v", string(cfg.Body))
	}
	if cfg.Rate != 42 {
		t.Errorf("Wrong rate")
	}
	if cfg.Header.Get("X-TYPE") != "Test" || cfg.Header.Get("X-SIZE") != "empty" {
		t.Errorf("Wrong header")
	}
}

var requestConfigPartial = []byte(`
resource: /hls/CH_UFCHD_HLS/master.m3u8
header:
    - name: X-TYPE
      value: Test
`)

func TestConfigRequestPartial(t *testing.T) {
	cfg := ConfigRequest{}

	if err := yaml.Unmarshal(requestConfigPartial, &cfg); err != nil {
		t.Errorf("Can't unmarsha cfg: %s", err.Error())
	}

	if cfg.Resource != "/hls/CH_UFCHD_HLS/master.m3u8" {
		t.Errorf("Wrong resource '%s'", cfg.Resource)
	}
	if cfg.Method != "GET" {
		t.Errorf("Wrong method '%s'", cfg.Method)
	}
	if cfg.Body != nil {
		t.Errorf("Wrong body %v", cfg.Body)
	}
	if cfg.Rate != 1 {
		t.Errorf("Wrong rate")
	}
	if cfg.Header.Get("X-TYPE") != "Test" {
		t.Errorf("Wrong header")
	}
}


var scriptConfig = []byte(`
name: 631e8c62-419e-46a2-b6cc-c3862c150843
header:
      - name: Content-Type
        value: text/m3u8
      - name: X-User-Token
        value: e04ce87c-74a9-4431-a829-4635f53583f3
requests:
      - resource: /enter
        method: POST
        body: VVNFUjpQQVNTV09SRA==
      - resource: /get/master.m3u8
        header:
          - name: Content-Type
            value: text/mediaplaylist
          - name: X-User-Token
            value: e04ce87c-74a9-4431-a829-4635f53583f3
      - resource: /get/media.m3u8
        rate: 100`)

func TestConfigScript(t *testing.T) {
	cfg := ConfigScript{}

	if e := yaml.Unmarshal(scriptConfig, &cfg); e != nil {
		t.Errorf("Failed to unmarshal: %s", e.Error())
	}

	if cfg.Name != "631e8c62-419e-46a2-b6cc-c3862c150843" {
		t.Errorf("Wromg script name")
	}

	if cfg.Header.Get("Content-Type") != "text/m3u8" ||
		cfg.Header.Get("X-User-Token") != "e04ce87c-74a9-4431-a829-4635f53583f3" {
		t.Errorf("Wrong headers")
	}

	if len(cfg.Requests) != 3 {
		t.Errorf("Wrong number of requests")
	}
}

var hostConfig = []byte(`
addr: 127.0.0.1
port: 8090
protocol: http`)

func TestConfigHost(t *testing.T) {
	cfg := ConfigHost{}

	if e := yaml.Unmarshal(hostConfig, &cfg); e != nil {
		t.Errorf("Failed to unmarshal %s", e.Error())
	}

	if cfg.Address != "127.0.0.1" {
		t.Errorf("Wrong host in config")
	}
	if cfg.Port != 8090 {
		t.Errorf("Wrong port number")
	}
	if cfg.Protocol != "http" {
		t.Errorf("Wrong prtocol")
	}
}

// TODO: Здесь должен быть тест на проверку конфигурации ядра. Но пока
// не понятно что еще я добалю в параметры, поэтому не буду его
// добавлять

func TestMergeHttpHedaers(t *testing.T) {
	alpha := make(http.Header)
	betta := make(http.Header)
	gamma := make(http.Header)

	alpha.Set("Content-Type", "bytes/octetstream")
	alpha.Set("Allow", "GET,OPTIONS,HEAD")
	alpha.Set("X-BEST-BEFORE", "today")

	betta.Set("Content-Type", "bytes/octetstream")
	betta.Set("Allow", "POST")

	gamma.Set("Content-Type", "text/json")

	final := MergeHttpHeaders(alpha, betta, gamma)
	if final.Get("Content-Type") != "text/json" {
		t.Errorf("Wrong Content-Type: %s", final.Get("Content-Type"))
	}
	if final.Get("Allow") != "POST" {
		t.Errorf("Wrong Allow: %s", final.Get("Allow"))
	}
	if final.Get("X-BEST-BEFORE") != "today" {
		t.Errorf("Wrong X-BEST-BEFORE: %s", final.Get("X-BEST-BEFORE"))
	}
}

func TestConfigRequestToURL(t *testing.T) {
	h80 := &ConfigHost{"127.0.0.1", 80, "http", nil}
	h443 := &ConfigHost{"localhost", 443, "https", nil}
	h8080 := &ConfigHost{"localhost", 8080, "https", nil}
	r1 := &ConfigRequest{"/someresid", "GET", nil, 10, nil}

	url80 := h80.GetUrlFor(r1)
	if url80 != "http://127.0.0.1/someresid" {
		t.Errorf("Wrong http url: %s", url80)
	}
	url443 := h443.GetUrlFor(r1)
	if url443 != "https://localhost/someresid" {
		t.Errorf("Wrong https url: %s", url443)
	}
	url8080 := h8080.GetUrlFor(r1)
	if url8080 != "https://localhost:8080/someresid" {
		t.Errorf("Wrong https on port 8080: %s", url8080)
	}
}
