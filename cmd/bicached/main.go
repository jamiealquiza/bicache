package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/jamiealquiza/bicache"
)

type Request struct {
	command string
	params  string
}

var (
	commands = map[string]func(c *bicache.Bicache, r *Request) string{
		"get": get,
		"set": set,
		"del": del,
	}
)

func main() {
	address := "localhost:9090"

	cache := bicache.New(&bicache.Config{
		MfuSize:   10,
		MruSize:   10,
		AutoEvict: 1000,
		EvictLog:  true,
	})

	go func(c *bicache.Bicache) {
		interval := time.NewTicker(time.Second * 5)
		defer interval.Stop()

		for _ = range interval.C {
			stats := c.Stats()
			j, _ := json.Marshal(stats)
			log.Println(string(j))
		}
	}(cache)

	log.Printf("Bicached started: %s\n", address)

	server, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Listener error: %s\n", err)
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("API error: %s\n", err)
			continue
		}
		reqHandler(cache, conn)
	}
}

func reqHandler(c *bicache.Bicache, conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	buf, err := reader.ReadBytes('\n')
	if err != nil {
		log.Printf("req error: %s\n", err)
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

func get(c *bicache.Bicache, r *Request) string {
	v := c.Get(r.params)
	if v != nil {
		return fmt.Sprintf("%s\n", v.(string))
	}

	return "nil\n"
}

func set(c *bicache.Bicache, r *Request) string {
	parts := strings.Split(r.params, ":")
	k, v := parts[0], parts[1]
	c.Set(k, v)

	return "ok\n"
}

func del(c *bicache.Bicache, r *Request) string {
	c.Del(r.params)

	return "ok\n"
}
