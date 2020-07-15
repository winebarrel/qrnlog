package qrnlog

import (
	"bufio"
	"io"
	"time"

	"github.com/jamiealquiza/tachymeter"
	"github.com/percona/go-mysql/query"

	jsoniter "github.com/json-iterator/go"
)

const PtFingerprint = "pt-fingerprint"
const ReadLineBufSize = 4096

type JsonLine struct {
	Query string        `json:"query"`
	Time  time.Duration `json:"time"`
}

type QueryData struct {
	Times     []time.Duration
	LastQuery string
	QuerySet  map[string]struct{}
}

type Metrics struct {
	Metrics     *tachymeter.Metrics
	LastQuery   string
	UniqueCount int
	TimePct     float64
}

func Normalize(file io.Reader) (map[string]*Metrics, error) {
	reader := bufio.NewReader(file)
	m := map[string]*QueryData{}

	for {
		line, err := readLine(reader)

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		var jl JsonLine

		if err := jsoniter.Unmarshal(line, &jl); err != nil {
			return nil, err
		}

		addQueryData(m, &jl)
	}

	return calculate(m), nil
}

func addQueryData(m map[string]*QueryData, jl *JsonLine) {
	fingerprint := query.Fingerprint(jl.Query)
	qd, ok := m[fingerprint]

	if !ok {
		qd = &QueryData{
			Times:    []time.Duration{},
			QuerySet: map[string]struct{}{},
		}
	}

	qd.Times = append(qd.Times, jl.Time)
	qd.LastQuery = jl.Query
	qd.QuerySet[jl.Query] = struct{}{}
	m[fingerprint] = qd
}

func calculate(m map[string]*QueryData) map[string]*Metrics {
	metricsByQuery := map[string]*Metrics{}
	tolalTime := time.Duration(0)

	for query, qd := range m {
		t := tachymeter.New(&tachymeter.Config{Size: len(qd.Times)})

		for _, tm := range qd.Times {
			t.AddTime(tm)
		}

		metrics := &Metrics{
			Metrics:     t.Calc(),
			LastQuery:   qd.LastQuery,
			UniqueCount: len(qd.QuerySet),
		}

		metricsByQuery[query] = metrics
		tolalTime += metrics.Metrics.Time.Cumulative
	}

	for _, m := range metricsByQuery {
		m.TimePct = float64(m.Metrics.Time.Cumulative) / float64(tolalTime) * 100
	}

	return metricsByQuery
}

func readLine(reader *bufio.Reader) ([]byte, error) {
	buf := make([]byte, 0, ReadLineBufSize)
	var err error

	for {
		line, isPrefix, e := reader.ReadLine()
		err = e

		if len(line) > 0 {
			buf = append(buf, line...)
		}

		if !isPrefix || err != nil {
			break
		}
	}

	return buf, err
}
