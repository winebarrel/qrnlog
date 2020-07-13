# qrnlog

qrnlog is a [qrn](https://github.com/winebarrel/qrn) log aggregation tool.

## Usage

```
usage: qrnlog [-help|-version] QRN_LOG
```

```
$ cat data.jsonl
{"query":"select 1"}
{"query":"select now()"}

$ qrn -dsn root:@/test -data data.jsonl -log log.jsonl
01:00 | 1 agents / run 904907 queries (14715 qps)

{
  "DSN": "root:@/test",
  ...
}

$ head log.jsonl
{"query":"select 1","time":1153591}
{"query":"select now()","time":160093}
{"query":"select 1","time":82081}
...

$ qrnlog log.jsonl
{"Count":460086,"Time":{"Cumulative":26670596278,"HMean":54857,"Avg":57968,"P50":53486,"P75":60209,"P95":91713,"P99":124969,"P999":175905,"Long5p":113644,"Short5p":40984,"Max":1153591,"Min":32122,"StdDev":17490,"Range":1121469},"Query":"select ?"}
{"Query":"select now()","Count":460086,"Time":{"Cumulative":27068202632,"HMean":56075,"Avg":58832,"P50":54373,"P75":60430,"P95":91866,"P99":126844,"P999":179187,"Long5p":115363,"Short5p":43446,"Max":1391779,"Min":33477,"StdDev":16908,"Range":1358302}}
```
