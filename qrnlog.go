package qrnlog

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jamiealquiza/tachymeter"
	jsoniter "github.com/json-iterator/go"
)

const PtFingerprint = "pt-fingerprint"

type QueryLog struct {
	Query string        `json:"query"`
	Time  time.Duration `json:"time"`
}

func Normalize(file io.Reader) (map[string]*tachymeter.Metrics, error) {
	qs, tms, err := parseLines(file)

	if err != nil {
		return nil, err
	}

	cmd, stdin, stdout, stderr, err := makeCmd(PtFingerprint)

	if err != nil {
		return nil, err
	}

	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = cmd.Process.Kill()
	}()

	// tailf stderr
	go func() {
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			fmt.Fprint(os.Stderr, scanner.Text())
		}
	}()

	done := make(chan map[string][]time.Duration)
	go aggregate(stdout, tms, done)

	for _, q := range qs {
		fmt.Fprintf(stdin, "%s;\n", q)
	}

	stdin.Close()
	m := <-done

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return calculate(m), nil
}

func parseLines(file io.Reader) ([]string, []time.Duration, error) {
	reader := bufio.NewReader(file)
	qs := []string{}
	tms := []time.Duration{}

	for {
		line, err := readLine(reader)

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}

		var queryLog QueryLog

		if err := jsoniter.Unmarshal(line, &queryLog); err != nil {
			return nil, nil, err
		}

		qs = append(qs, queryLog.Query)
		tms = append(tms, queryLog.Time)
	}

	return qs, tms, nil
}

func aggregate(reader io.Reader, tms []time.Duration, done chan map[string][]time.Duration) {
	defer func() {
		close(done)
	}()

	m := map[string][]time.Duration{}
	scanner := bufio.NewScanner(reader)

	for i := 0; scanner.Scan(); i++ {
		query := scanner.Text()
		ts, ok := m[query]

		if !ok {
			ts = []time.Duration{}
		}

		ts = append(ts, tms[i])
		m[query] = ts
	}

	done <- m
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
