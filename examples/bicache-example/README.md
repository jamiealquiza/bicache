Excerpts:
```go
c, _ := bicache.New(&bicache.Config{
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
Cumulative:	19.242196ms
Avg.:		192ns
p50: 		182ns
p75:		219ns
p95:		311ns
p99:		446ns
p999:		589ns
Long 5%:	481ns
Short 5%:	100ns
Max:		281.841µs
Min:		82ns
Rate/sec.:	5196912.04

{"MfuSize":0,"MruSize":100000,"MfuUsedP":0,"MruUsedP":200,"Hits":100010,"Misses":0,"Evictions":0}
```

Note: timing data via [tachymeter](https://github.com/jamiealquiza/tachymeter)