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

func Normalize(file io.Reader) (map[string]*tachymeter.Metrics, error) {
	reader := bufio.NewReader(file)
	m := map[string][]time.Duration{}

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
		ts, ok := m[fingerprint]

		if !ok {
			ts = []time.Duration{}
		}

		ts = append(ts, queryLog.Time)
		m[fingerprint] = ts
	}

	return calculate(m), nil
}

func calculate(m map[string][]time.Duration) map[string]*tachymeter.Metrics {
	metricsByQuery := map[string]*tachymeter.Metrics{}

	for query, ts := range m {
		t := tachymeter.New(&tachymeter.Config{Size: len(ts)})

		for _, tm := range ts {
			t.AddTime(tm)
		}

		metricsByQuery[query] = t.Calc()
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
