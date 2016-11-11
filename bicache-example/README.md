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
50000 samples of 50000 events
Total:			49.690964ms
Avg.:			993ns
Median: 		450ns
95%ile:			2.496µs
Longest 5%:		8.577µs
Shortest 5%:	371ns
Max:			1.517224ms
Min:			347ns
Rate/sec.:		1006219.16

[ Get 100000 keys ]
50000 samples of 50000 events
Total:			11.500957ms
Avg.:			230ns
Median: 		216ns
95%ile:			301ns
Longest 5%:		388ns
Shortest 5%:	189ns
Max:			13.697µs
Min:			168ns
Rate/sec.:		4347464.30

{"MfuSize":25000,"MruSize":25000,"MfuUsedP":50,"MruUsedP":100}
```
