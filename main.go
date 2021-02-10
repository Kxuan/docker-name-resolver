package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/miekg/dns"
	"log"
	"os"
	"strings"
)

var docker *client.Client
var ctx context.Context

func resolveName(name string) (result string, err error) {
	firstdot := strings.Index(name, ".")
	if firstdot < 0 {
		return "", fmt.Errorf("invalid name")
	}
	id := name[0:firstdot]
	r, err := docker.ContainerInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("not found")
	}
	if r.NetworkSettings == nil {
		return "", fmt.Errorf("no network settings")
	}
	result = r.NetworkSettings.IPAddress
	return
}

func handleQueryA(m *dns.Msg, q *dns.Question) {
	log.Printf("Query for %s\n", q.Name)
	result, err := resolveName(q.Name)
	if err != nil {
		return
	}

	rr, err := dns.NewRR(fmt.Sprintf("%s 1 A %s", q.Name, result))
	if err == nil {
		m.Answer = append(m.Answer, rr)
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode != dns.OpcodeQuery {
		m.Rcode = dns.RcodeNotImplemented
	} else {
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				handleQueryA(m, &q)
			}
		}
	}

	w.WriteMsg(m)
}

func main() {
	var err error

	ctx = context.Background()
	docker, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	dns.HandleFunc(".", handleDnsRequest)

	addr, got := os.LookupEnv("BIND_ADDRESS")
	if !got {
		addr = "127.0.0.11"
	}
	bind := fmt.Sprintf("%s:%d", addr, 53)
	server := &dns.Server{Addr: bind, Net: "udp"}
	log.Printf("Starting at %s\n", bind)
	err = server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
