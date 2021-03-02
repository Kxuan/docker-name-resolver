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

func resolveName(name string) (result []string, err error) {
	firstdot := strings.Index(name, ".")
	if firstdot < 0 {
		return nil, fmt.Errorf("invalid name")
	}
	id := name[0:firstdot]

	r, err := docker.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("not found, %v", err)
	}
	if r.NetworkSettings == nil {
		return nil, fmt.Errorf("no network settings")
	}
	if r.NetworkSettings.IPAddress != "" {
		result = append(result, r.NetworkSettings.IPAddress)
	}
	for _, bridge := range r.NetworkSettings.Networks {
		if bridge.IPAddress != "" {
			result = append(result, bridge.IPAddress)
		}
	}
	return
}

func handleQueryA(m *dns.Msg, q *dns.Question) {
	result, err := resolveName(q.Name)
	if err != nil {
		log.Printf("%v: %v\n", q.Name, err)
		return
	}

	for _, ip := range result {
		rr, err := dns.NewRR(fmt.Sprintf("%s 1 A %s", q.Name, ip))
		if err == nil {
			m.Answer = append(m.Answer, rr)
		} else {
			m.Rcode = dns.RcodeRefused
		}
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
		/**
		  Since we do not recursive search internet domain name, we have to answer some error code to DNS client.
		  If we just replied normally, the DNS client may given up, and failed too early.
		*/
		if len(m.Answer) == 0 {
			m.Rcode = dns.RcodeRefused
		}
	}

	_ = w.WriteMsg(m)
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
	udpSvr := &dns.Server{Addr: bind, Net: "udp"}
	tcpSvr := &dns.Server{Addr: bind, Net: "tcp"}
	log.Printf("Starting at %s\n", bind)

	ch := make(chan error)
	go func() {
		err := udpSvr.ListenAndServe()
		defer udpSvr.Shutdown()
		ch <- err
	}()
	go func() {
		err := tcpSvr.ListenAndServe()
		defer tcpSvr.Shutdown()
		ch <- err
	}()
	err = <-ch
	if err != nil {
		log.Fatalf("Failed to start udp_svr: %s\n ", err.Error())
	}
}
