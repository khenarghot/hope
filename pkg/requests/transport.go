package requests

import (
	"crypto/tls"
	op "github.com/khenarghot/hope/pkg/options"
	"golang.org/x/net/http2"
	"net/http"
)

const (
	// TransportHTTP — configure HTTP transport
	TransportHTTP = iota
	// TransportHTTP2 — configure HTTP2 transport
	TransportHTTP2
)

type TransportParametrs struct {
	Host            string
	IdleConnections int
}

func GetNewTransport(transport int, p interface{}) http.RoundTripper {

	tp, ok := p.(TransportParametrs)
	if !ok {
		panic("This is small piece of shit — unimplimented HTTP3 support")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         tp.Host,
		},
		MaxIdleConnsPerHost: min(tp.IdleConnections,
			op.HopeConfig.Core.Connections),
		DisableCompression: true,
		DisableKeepAlives:  true,
		Proxy:              nil,
	}

	switch transport {
	case TransportHTTP:
		tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	case TransportHTTP2:
		http2.ConfigureTransport(tr)
	default:
		panic("Not implimented transport")
	}
	return tr
}
