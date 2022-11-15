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

	//"sync"

	ping "github.com/mbia-ITU/DISYS-HandIn-4/gRPC/gRPC"
	"google.golang.org/grpc"
)

var i_want_to_get_into_citicalsection bool
var queued_requests []int32

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
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		p.sendPingToAll()
	}
}

type peer struct {
	ping.UnimplementedPingServer
	id        int32
	timestamp int32
	clients   map[int32]ping.PingClient
	ctx       context.Context
}

func (p *peer) Ping(ctx context.Context, req *ping.Request) (*ping.Reply, error) {
	if i_want_to_get_into_citicalsection && send_priority_to_all(p, req.GetId(), req.GetTimestamp()) {
		time.Sleep(10 * time.Second)
		add_to_request_queue(req.GetId())
		return nil, nil
		//Only beat one other peer. How do we make it wait until other peers are checked and beaten?
	} else {
		rep := &ping.Reply{Message: "you're free to go"}
		p.timestamp++
		return rep, nil
	}
}

func (p *peer) sendPingToAll() {
	request := &ping.Request{Id: p.id, Timestamp: p.timestamp}
	i_want_to_get_into_citicalsection = true
	for id, client := range p.clients {
		reply, err := client.Ping(p.ctx, request)
		if err != nil {
			fmt.Println("something went wrong")
		}
		fmt.Printf("Got reply from id %v: %v\n", id, reply.Message)
		p.timestamp++
	}
	/*Access critical section*/

}

func send_priority_to_all(p *peer, id int32, clock int32) bool {
	//Send ping.priority to all peeras and use waitgroup(wg from "sync") to wait for response for every peer
	if p.timestamp > clock {
		return true
	} else if clock > p.timestamp {
		return false
	} else {
		if p.id < id {
			return true
		} else {
			return false
		}
	}
}

func add_to_request_queue(id int32) {
	queued_requests = append(queued_requests, id)
}
