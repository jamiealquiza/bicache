package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/jamiealquiza/bicache"
)

// Request holds an API request
// command and parameters.
type Request struct {
	command string
	params  string
}

var (
	// Commands is a map of valid API requests
	// to internal functions.
	commands = map[string]func(c *bicache.Bicache, r *Request) string{
		"get": get,
		"set": set,
		"del": del,
	}
)

func main() {
	// Initialize settings.
	var address = flag.String("listen", "localhost:9090", "listen address")
	var mfuSize = flag.Uint("mfu", 256, "MFU cache size")
	var mruSize = flag.Uint("mru", 64, "MRU cache size")
	var evictInterval = flag.Uint("evict-interval", 1000, "Eviction interval in ms")
	var evictLog = flag.Bool("evict-log", true, "log eviction times")
	flag.Parse()

	// Instantiate Bicache.
	cache := bicache.New(&bicache.Config{
		MfuSize:   *mfuSize,
		MruSize:   *mruSize,
		AutoEvict: *evictInterval,
		EvictLog:  *evictLog,
	})

	// Log Bicache stats on interval.
	go func(c *bicache.Bicache) {
		interval := time.NewTicker(time.Second * 5)
		defer interval.Stop()

		for _ = range interval.C {
			stats := c.Stats()
			j, _ := json.Marshal(stats)
			log.Println(string(j))
		}
	}(cache)

	// Setup the TCP listener.
	server, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	log.Printf("Bicached Listening: %s\n", *address)

	// Request listener loop.
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("req error: %s\n", err)
			continue
		}
		reqHandler(cache, conn)
	}
}

// Request handler takes API requests
// and passes them to the appropriate bicache
// method.
func reqHandler(c *bicache.Bicache, conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	buf, err := reader.ReadBytes('\n')
	if err != nil {
		// TODO might want an explicit
		// action here.
		return
	}

	input := strings.Fields(string(buf[:len(buf)-1]))
	request := &Request{command: input[0]}

	if len(input) > 1 {
		request.params = input[1]
	}

	if command, valid := commands[request.command]; valid {
		response := command(c, request)
		conn.Write([]byte(response))
	} else {
		m := fmt.Sprintf("not a command: %s\n", request.command)
		conn.Write([]byte(m))
	}
}

// Bicache Get method.
func get(c *bicache.Bicache, r *Request) string {
	v := c.Get(r.params)
	if v != nil {
		return fmt.Sprintf("%s\n", v.(string))
	}

	return "nil\n"
}

// Bicache Set method.
func set(c *bicache.Bicache, r *Request) string {
	parts := strings.Split(r.params, ":")
	k, v := parts[0], parts[1]
	c.Set(k, v)

	return "ok\n"
}

// Bicache Del method.
func del(c *bicache.Bicache, r *Request) string {
	c.Del(r.params)

	return "ok\n"
}
