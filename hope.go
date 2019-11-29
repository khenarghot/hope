package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/khenarghot/hope/pkg/options"
	"github.com/khenarghot/hope/pkg/requests"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s: \n", os.Args[0])
		fmt.Fprintf(out, "\t%s load.yaml\n", os.Args[0])
		flag.PrintDefaults()
	}

	// TODO: Добавить опции для генерации дефолного конфига и прочей билиберды
	var timeFlag duration = time.Duration{0}
	workers := flag.Int("workers", 0, "number of clients")
	reqNum := flag.Int("requests", 0, "number of requests pass to the clent")
	flag.Val(&duration, "d", "total test duration")

	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments %d\n", len(flag.Args()))
		flag.Usage()
		os.Exit(1)
	}

	if e := options.LoadConfigFromFile(flag.Args()[0]); e != nil {
		fmt.Fprintf(os.Stderr, "Failed parse config file: %s\n", e.Error())
		os.Exit(2)
	}
	if workers > 0 {
		options.HopeConfig.Core.Workers = workers
	}
	if reqNum != 0 {
		if reqNum < 0 {
			options.HopeConfig.Core.Requests = 0
		}
		options.HopeConfig.Core.Requests = reqNum
	}
	if duration != time.Duration(0) {
		options.HopeConfig.Core.Duration = duration
	}

	fmt.Println(options.HopeConfig.Core, options.HopeConfig.Host, options.HopeConfig.Scripts[0])

	if options.HopeConfig.Core.Workers == 0 {
		fmt.Fprintf(os.Stderr, "Need some workers in core:workers\n")
		os.Exit(2)
	}

	if options.HopeConfig.Core.Requests == 0 {
		fmt.Fprintf(os.Stderr, "Need some connections in core:connections\n")
		os.Exit(2)
	}

	transport, err := CompileTransport(options.HopeConfig.Host, options.HopeConfig.Core)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed configure transport: %s\n", err.Error())
		os.Exit(2)
	}

	req, err := CompileRequests(options.HopeConfig.Host, options.HopeConfig.Scripts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed generate scripts: %s\n", err.Error())
		os.Exit(2)
	}

	task := requests.NewTask(requests.NewDefaultCollectot(), transport,
		req, 0, options.HopeConfig.Core.Workers, options.HopeConfig.Core.Requests, time.Second)

	task.Init()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		task.Stop()
	}()
	if options.HopeConfig.Core.Duration > 0 {
		go func() {
			time.Sleep(options.HopeConfig.Core.Duration)
			task.Stop()
		}()
	}
	task.Run()
}

func CompileRequests(host *options.ConfigHost,
	scripts []*options.ConfigScript) ([][]*requests.Request, error) {
	res := make([][]*requests.Request, len(scripts))
	for i, script := range scripts {
		cs := make([]*requests.Request, len(script.Requests))
		for j, req := range script.Requests {
			header := options.MergeHttpHeaders(host.Header, script.Header, req.Header)
			ireq, e := requests.NewRequest(host.GetUrlFor(&req),
				req.Method, header, req.Body, req.Rate)
			if e != nil {
				return nil, e
			}
			cs[j] = ireq
		}
		res[i] = cs
	}
	return res, nil
}

func CompileTransport(host *options.ConfigHost, core *options.ConfigCore) (http.RoundTripper, error) {
	var transport int

	switch host.Protocol {
	case "HTTP", "HTTPS":
		transport = requests.TransportHTTP
	case "HTTP2", "HTTPS2":
		transport = requests.TransportHTTP2
	default:
		return nil, fmt.Errorf("Wrong protocol: %s", host.Protocol)
	}

	return requests.GetNewTransport(transport,
		requests.TransportParametrs{host.Address, core.Workers}), nil
}

type timeFlag struct {
	time.Duration
}

func (t *timeFlag) String() {
	return t.Duration.String()
}

func (t *timeFlag) Set(s string) (e error) {
	t.Duration, e = time.ParseDuration(s)
	return e
}
