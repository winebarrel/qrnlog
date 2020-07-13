package qrnlog

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/jamiealquiza/tachymeter"
	jsoniter "github.com/json-iterator/go"
)

var ReadLineBufSize = 4096

const PtFingerprint = "pt-fingerprint"

type QueryLog struct {
	Query string        `json:"query"`
	Time  time.Duration `json:"time"`
}

func Normalize(file io.Reader) (map[string]*tachymeter.Metrics, error) {
	cmd, stdin, stdout, stderr, err := makeCmd(PtFingerprint)

	if err != nil {
		return nil, err
	}

	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	ptFingerprint := make(chan time.Duration)
	m := &sync.Map{}

	go tailfStderr(stderr)
	go aggregate(stdout, ptFingerprint, m)

	reader := bufio.NewReader(file)

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

		fmt.Fprintf(stdin, "%s;\n", queryLog.Query)
		ptFingerprint <- queryLog.Time
	}

	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return calculate(m), nil
}

func makeCmd(s string) (cmd *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, stderr io.ReadCloser, err error) {
	cmd = exec.Command(s)

	stdin, err = cmd.StdinPipe()

	if err != nil {
		return
	}

	stdout, err = cmd.StdoutPipe()

	if err != nil {
		return
	}

	stderr, err = cmd.StderrPipe()

	if err != nil {
		return
	}

	return
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

func tailfStderr(reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		fmt.Fprint(os.Stderr, scanner.Text())
	}
}

func aggregate(reader io.Reader, c chan time.Duration, m *sync.Map) {
	scanner := bufio.NewScanner(reader)

	for tm := range c {
		if scanner.Scan() {
			query := scanner.Text()
			v, ok := m.Load(query)
			var ts []time.Duration

			if ok {
				ts = v.([]time.Duration)
			} else {
				ts = []time.Duration{}
			}

			ts = append(ts, tm)
			m.Store(query, ts)
		} else {
			log.Fatalf("cannot read line from log")
		}
	}
}

func calculate(m *sync.Map) map[string]*tachymeter.Metrics {
	metricsByQuery := map[string]*tachymeter.Metrics{}

	m.Range(func(k, v interface{}) bool {
		query := k.(string)
		ts := v.([]time.Duration)
		t := tachymeter.New(&tachymeter.Config{Size: len(ts)})

		for _, tm := range ts {
			t.AddTime(tm)
		}

		metricsByQuery[query] = t.Calc()

		return true
	})

	return metricsByQuery
}
