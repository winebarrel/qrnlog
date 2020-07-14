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

type Metrics struct {
	Metrics     *tachymeter.Metrics
	Query       string
	UniqueCount int
}

func Normalize(file io.Reader) (map[string]*Metrics, error) {
	reader := bufio.NewReader(file)
	m := map[string]*TimesQuery{}
	uq := map[string]map[string]struct{}{}

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
		addTimesQuery(m, fingerprint, &queryLog)
		addUniqueCount(uq, fingerprint, queryLog.Query)
	}

	return calculate(m, uq), nil
}

func addTimesQuery(m map[string]*TimesQuery, fingerprint string, ql *QueryLog) {
	tsq, ok := m[fingerprint]

	if !ok {
		tsq = &TimesQuery{
			Times: []time.Duration{},
			Query: ql.Query,
		}
	}

	tsq.Times = append(tsq.Times, ql.Time)
	m[fingerprint] = tsq
}

func addUniqueCount(m map[string]map[string]struct{}, fingerprint string, query string) {
	xByQuery, ok := m[fingerprint]

	if !ok {
		xByQuery = map[string]struct{}{}
	}

	xByQuery[query] = struct{}{}
	m[fingerprint] = xByQuery
}

func calculate(m map[string]*TimesQuery, uq map[string]map[string]struct{}) map[string]*Metrics {
	metricsByQuery := map[string]*Metrics{}

	for query, tsq := range m {
		t := tachymeter.New(&tachymeter.Config{Size: len(tsq.Times)})

		for _, tm := range tsq.Times {
			t.AddTime(tm)
		}

		metricsByQuery[query] = &Metrics{
			Metrics:     t.Calc(),
			Query:       tsq.Query,
			UniqueCount: len(uq[query]),
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
