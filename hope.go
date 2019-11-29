package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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
	var duration timeFlag = timeFlag{time.Duration(0)}
	workers := flag.Int("workers", 0, "number of clients")
	reqNum := flag.Int("requests", 0, "number of requests pass to the clent")
	mktemplate := flag.Bool("generate", false, "generate template file for load")
	flag.Var(&duration, "d", "total test duration")
	flag.Parse()

	if *mktemplate {
		if len(flag.Args()) > 0 {
			if e := ioutil.WriteFile(flag.Args()[0], loadTemplate, 0644); e != nil {
				fmt.Fprintf(os.Stderr, "Failed write template to %s: %s",
					flag.Args()[0], e.Error())
				os.Exit(1)
			}
		} else {
			os.Stdout.Write(loadTemplate)
		}
		os.Exit(0)
	}

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Wrong number of arguments %d\n", len(flag.Args()))
		flag.Usage()
		os.Exit(1)
	}
	if e := options.LoadConfigFromFile(flag.Args()[0]); e != nil {
		fmt.Fprintf(os.Stderr, "Failed parse config file: %s\n", e.Error())
		os.Exit(2)
	}
	if *workers > 0 {
		options.HopeConfig.Core.Workers = *workers
	}
	if *reqNum != 0 {
		if *reqNum < 0 {
			options.HopeConfig.Core.Requests = 0
		}
		options.HopeConfig.Core.Requests = *reqNum
	}
	if duration.Duration != time.Duration(0) {
		options.HopeConfig.Core.Duration = duration.Duration
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

func (t *timeFlag) String() string {
	return t.Duration.String()
}

func (t *timeFlag) Set(s string) (e error) {
	t.Duration, e = time.ParseDuration(s)
	return e
}

var loadTemplate []byte = []byte(`
# Бзовые настройки. Могут быть опредеены через командную строку
core:
  # Число одновременных запросов должно быть определено или тут или в
  # командной строке
  workers: 80

  # Общее число запросов, после которого нужно остановиться. Если 0 то
  # не остановиться никогда
  #requests: 1000

  # Длительность теста. Тест завершитьс по прошествии этого времени,
  # даже если не все запросывыполнены. Если не устанавливать то
  # завершиться лишь по исполнении всех запросов.
  #duration: 5m10s

# Общая настройка узла
host:
  # Адрес испытываемого узла (Обязателен)
  addr: 127.0.0.1

  # Используемы порт (80 по умолчанию)
  #port: 80

  # Потокол (HTTP/HTTP2/HTTPS/HTTPS2) (HTTP по умолчанию)
  #protocol: HTTP

  # Для хоста может быть определен заголовк как и для любой другой части
  #header:
  #  - name: Content-Type
  #    value: text/mediaplaylist
  #  - name: X-User-Token
  #    value: e04ce87c-74a9-4431-a829-4635f53583f3

# Сценари для исполннения нагрузки
scripts:
  - name: af5e6e9b-a626-45b3-8f7f-45f999531935
    header:
      - name: X-User-Token
	value: e04ce87c-74a9-4431-a829-4635f53583f3
    requests:
      - resource: /auth
	# Метод для реусурса. по умолчанию Get
	method: POST
	# base64 кодированное тело запроса
	body: VVNFUjpQQVNTV09SRA==
	header:
	- name: Content-Type
	  value: text/plain
      - resource: /play/channel/variant.m3u8
      - resource: /play/channel/playlist.m3u8


`)
