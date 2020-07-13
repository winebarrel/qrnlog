package qrnlog

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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

	defer func() {
		_ = cmd.Process.Kill()
	}()

	ch := make(chan time.Duration)
	done := make(chan map[string][]time.Duration)

	go tailfStderr(stderr)
	go aggregate(stdout, ch, done)

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
		ch <- queryLog.Time
	}

	close(ch)
	stdin.Close()
	m := <-done

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

func aggregate(reader io.Reader, ch chan time.Duration, done chan map[string][]time.Duration) {
	defer func() {
		close(done)
	}()

	m := map[string][]time.Duration{}
	scanner := bufio.NewScanner(reader)

	for tm := range ch {
		if scanner.Scan() {
			query := scanner.Text()
			ts, ok := m[query]

			if !ok {
				ts = []time.Duration{}
			}

			ts = append(ts, tm)
			m[query] = ts
		} else {
			log.Fatalf("cannot read line from log")
		}
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
