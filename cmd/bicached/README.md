# Bicached

Bicache server.

# Install
Tested with Go 1.7.

- `go get -u github.com/jamiealquiza/bicache`
- `go install github.com/jamiealquiza/bicache/cmd/bicached`

# Usage
```
bicached -h
Usage of bicached:
  -evict-interval uint
        Eviction interval in ms (default 1000)
  -evict-log
        log eviction times (default true)
  -listen string
        listen address (default "localhost:9090")
  -mfu uint
        MFU cache size (default 256)
  -mru uint
        MRU cache size (default 64)
```

# Examples

Set, Get 10 keys:
```
% for i in 1 2 3 4 5 6 7 8 9 10; do echo "set $i:value" | nc localhost 9090; done
ok
ok
ok
ok
ok
ok
ok
ok
ok
ok
% for i in 1 2 3 4 5 6 7 8 9 10; do echo "get $i" | nc localhost 9090; done
value
value
value
value
value
value
value
value
value
value
```

Bicached Server:
```
% bicached -evict-interval=10000 -mru=5
2016/12/04 10:28:20 Bicached started: localhost:9090
2016/12/04 10:28:25 {"MfuSize":0,"MruSize":10,"MfuUsedP":0,"MruUsedP":200}
2016/12/04 10:28:30 AutoEvict ran in 23.935µs
2016/12/04 10:28:30 {"MfuSize":5,"MruSize":5,"MfuUsedP":1,"MruUsedP":100}
2016/12/04 10:28:35 {"MfuSize":5,"MruSize":5,"MfuUsedP":1,"MruUsedP":100}
2016/12/04 10:28:40 {"MfuSize":5,"MruSize":5,"MfuUsedP":1,"MruUsedP":100}
2016/12/04 10:28:40 AutoEvict ran in 12.563µs
```
