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

type QueryLog struct {
	Query string        `json:"query"`
	Time  time.Duration `json:"time"`
}

type TimesQuery struct {
	Times []time.Duration
	Query string
}

type MetricsQuery struct {
	Query   string
	Metrics *tachymeter.Metrics
}

func Normalize(file io.Reader) (map[string]*MetricsQuery, error) {
	reader := bufio.NewReader(file)
	m := map[string]*TimesQuery{}

	for {
		line, err := readLine(reader)

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		var queryLog QueryLog

		if err := jsoniter.Unmarshal(line, &queryLog); err != nil {
			return nil, err
		}

		fingerprint := query.Fingerprint(queryLog.Query)
		tsq, ok := m[fingerprint]

		if !ok {
			tsq = &TimesQuery{
				Times: []time.Duration{},
				Query: queryLog.Query,
			}
		}

		tsq.Times = append(tsq.Times, queryLog.Time)
		m[fingerprint] = tsq
	}

	return calculate(m), nil
}

func calculate(m map[string]*TimesQuery) map[string]*MetricsQuery {
	metricsByQuery := map[string]*MetricsQuery{}

	for query, tsq := range m {
		t := tachymeter.New(&tachymeter.Config{Size: len(tsq.Times)})

		for _, tm := range tsq.Times {
			t.AddTime(tm)
		}

		metricsByQuery[query] = &MetricsQuery{
			Metrics: t.Calc(),
			Query:   tsq.Query,
		}
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
