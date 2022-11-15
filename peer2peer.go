package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	ping "github.com/mbia-ITU/DISYS-HandIn-4/gRPC/gRPC"
	"google.golang.org/grpc"
)

var i_want_to_get_into_citicalsection bool

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &peer{
		id:        ownPort,
		timestamp: 0,
		clients:   make(map[int32]ping.PingClient),
		ctx:       ctx,
	}

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	grpcServer := grpc.NewServer()
	ping.RegisterPingServer(grpcServer, p)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)

		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := ping.NewPingClient(conn)
		p.clients[port] = c
		fmt.Printf("successfully dialed: %v\n", port)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		p.sendPingToAll()
	}
}

// Each peer is a struct with an id, timestamp, a list of connected peers and context
type peer struct {
	ping.UnimplementedPingServer
	id        int32
	timestamp int32
	clients   map[int32]ping.PingClient
	ctx       context.Context
}

// The Ping() function handles priority and check if the peer, who wants access to the critical section, can actually get access.
func (p *peer) Ping(ctx context.Context, req *ping.Request) (*ping.Reply, error) {
	if req.Timestamp > p.timestamp {
		p.timestamp = req.Timestamp
	}
	if i_want_to_get_into_citicalsection && req.Timestamp > p.timestamp {
		rep := &ping.Reply{Message: "I was first! (timestamp prio)"}
		p.timestamp++
		return rep, nil
	} else if i_want_to_get_into_citicalsection && req.Timestamp == p.timestamp && req.Id < p.id {
		rep := &ping.Reply{Message: "I was first! (id prio)"}
		p.timestamp++
		return rep, nil
	} else {
		rep := &ping.Reply{Message: "you're free to go"}
		p.timestamp++
		return rep, nil
	}
}

// The sendPingToAll() function is responsible for "asking" for access to the critical section and asks every other peer if it can get access. It also simulates the actual critical section.
func (p *peer) sendPingToAll() {
	request := &ping.Request{Id: p.id, Timestamp: p.timestamp}
	i_want_to_get_into_citicalsection = true
	done := false
	fmt.Println("Requesting access to critical section from other cliets...")
	for !done {
		done = true
		for id, client := range p.clients {
			reply, err := client.Ping(p.ctx, request)
			time.Sleep(4 * time.Second)
			if err != nil {
				fmt.Println("something went wrong")
			}
			if reply.Message == "I was first! (id prio)" {
				done = false
				fmt.Printf("Got reply from id %v: %v\n", id, reply.Message)
				p.timestamp++
				fmt.Println("Resending request...")
				break
			} else if reply.Message == "I was first! (timestamp prio)" {
				done = false
				fmt.Printf("Got reply from id %v: %v\n", id, reply.Message)
				p.timestamp++
				fmt.Println("Resending request...")
				break
			} else {
				fmt.Printf("Got reply from id %v: %v\n", id, reply.Message)
				p.timestamp++
			}
		}
	}
	fmt.Printf("%v just accessed the critical section at timestamp %v\n", p.id, p.timestamp)
	i_want_to_get_into_citicalsection = false
	p.timestamp++

}
