package requests

import (
	"bytes"
	"text/template"
	"time"
)

type Report struct {
	Count    int64         `yaml:"count"`
	Rps      int64         `yaml:"rps"`
	DataSize int64         `yaml:"total_data_recieved"`
	Duration time.Duration `yaml:"execution_time"`
	Start    time.Time     `yaml:"start_time"`

	RawDuration time.Duration `yaml:"requests_duration"`
	AvgDuration time.Duration `yaml:"requests_avg_duration"`
	AvgSize     int64         `yaml:"response_size"`
	Slowest     time.Duration `yaml:"slowest_request"`
	Fastest     time.Duration `yaml:"fastest_request"`

	OkResponse   int `yaml:"ok_response_count"`
	OverResponse int `yaml:"not_ok_response_count"`
	Errors       int `yaml:"errors_count"`
}

var (
	hrfTemplate = `
Summary:
  Requests:     {{ .Count }}
  Slowest:	{{ .Slowest }}
  Fastest:	{{ .Fastest }} 
  Average:	{{ .AvgDuration }} 
  Requests/sec:	{{  .Rps }}

  Total data:	{{ .DataSize }} bytes
  Size/request:	{{ .AvgSize }} bytes

  20x-30x responses:     {{ .OkResponse }}
  Non 20x-30x responses: {{ .OverResponse }}
 
  Errors: {{ .Errors }}
`
)

func (r *Report) String() string {
	buf := bytes.Buffer{}
	tmpl := template.Must(template.New("results").Parse(hrfTemplate))
	if e := tmpl.Execute(&buf, *r); e != nil {
		return e.Error()
	}
	return buf.String()
}
