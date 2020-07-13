package qrnlog

import (
	"bufio"
	"io"
	"os/exec"
)

var ReadLineBufSize = 4096

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
