# qrnlog

qrnlog is a [qrn](https://github.com/winebarrel/qrn) log normalization tool.

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

$ qrnlog log.jsonl # or `cat log.jsonl | qrnlog`
{"Query":"select now()","LastQuery":"select now()","Count":134118,"Time":{"Cumulative":9422205208,"HMean":68006,"Avg":70253,"P50":67195,"P75":72774,"P95":102030,"P99":133245,"P999":179611,"Long5p":123863,"Short5p":53248,"Max":1302663,"Min":36794,"StdDev":15983,"Range":1265869}}
{"Query":"select ?","LastQuery":"select 1","Count":134118,"Time":{"Cumulative":8655022244,"HMean":62221,"Avg":64532,"P50":61012,"P75":65928,"P95":98715,"P99":124093,"P999":166922,"Long5p":115808,"Short5p":48514,"Max":348380,"Min":34074,"StdDev":15021,"Range":314306}}
```
