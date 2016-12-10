Excerpts:
```go
c := bicache.New(&bicache.Config{
	MfuSize:   50000,
	MruSize:   25000,
	AutoEvict: 1000,
})

for i := 0; i < 50000; i++ {
	start := time.Now()
	c.Set(i, i*2)
}

for i := 0; i < 50000; i++ {
	_ = c.Get(i)
	t.AddTime(time.Since(start))
}
```


Running:
```
% bicache-example
[ Set 100000 keys ]
100000 samples of 100000 events
Cumulative:     102.444759ms
Avg.:           1.024µs
p50:            427ns
p75:            719ns
p95:            3.035µs
p99:            8.33µs
p999:           25.525µs
Long 5%:        9.198µs
Short 5%:       273ns
Max:            2.168331ms
Min:            248ns
Rate/sec.:      976135.83

[ Get 100000 keys ]
100000 samples of 100000 events
Cumulative:     21.709667ms
Avg.:           217ns
p50:            206ns
p75:            233ns
p95:            341ns
p99:            441ns
p999:           579ns
Long 5%:        443ns
Short 5%:       155ns
Max:            91.837µs
Min:            135ns
Rate/sec.:      4606242.92

{"MfuSize":25000,"MruSize":25000,"MfuUsedP":50,"MruUsedP":100}
```
