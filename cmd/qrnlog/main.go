package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"qrnlog"

	jsoniter "github.com/json-iterator/go"
)

var version string

type QueryTime struct {
	Query     string
	LastQuery string
	Count     int
	Time      interface{}
}

func init() {
	log.SetFlags(0)
}

func main() {
	file := parseArgs()
	defer file.Close()

	m, err := qrnlog.Normalize(file)

	if err != nil {
		log.Fatal(err)
	}

	for query, mq := range m {
		qt := QueryTime{
			Query:     query,
			LastQuery: mq.Query,
			Count:     mq.Metrics.Count,
			Time:      mq.Metrics.Time,
		}

		line, err := jsoniter.ConfigFastest.MarshalToString(qt)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(line)
	}
}

func parseArgs() io.ReadCloser {
	if len(os.Args) == 1 {
		return os.Stdin
	}

	if len(os.Args) > 2 || os.Args[1] == "-help" {
		log.Fatalf("usage: %s [-help|-version] QRN_LOG", os.Args[0])
	}

	if os.Args[1] == "-version" {
		fmt.Fprintln(os.Stderr, version)
		os.Exit(0)
	}

	file, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0)

	if err != nil {
		log.Fatal(err)
	}

	return file
}
