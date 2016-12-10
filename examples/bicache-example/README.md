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
Cumulative:     21.523946ms
Avg.:           215ns
p50:            191ns
p75:            215ns
p95:            359ns
p99:            495ns
p999:           2.811µs
Long 5%:        695ns
Short 5%:       137ns
Max:            49.386µs
Min:            121ns
Rate/sec.:      4645988.24

[ Get 100000 keys ]
100000 samples of 100000 events
Cumulative:     23.596428ms
Avg.:           235ns
p50:            218ns
p75:            241ns
p95:            400ns
p99:            566ns
p999:           1.272µs
Long 5%:        564ns
Short 5%:       151ns
Max:            35.76µs
Min:            135ns
Rate/sec.:      4237929.57

{"MfuSize":25000,"MruSize":25000,"MfuUsedP":50,"MruUsedP":100}
```

Note: timing data via [tachymeter](https://github.com/jamiealquiza/tachymeter)