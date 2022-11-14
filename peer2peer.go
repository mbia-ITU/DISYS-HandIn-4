package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	ping "github.com/mbia-ITU/DISYS-HandIn-4/gRPC/gRPC"
	"google.golang.org/grpc"
)

var citicalsection bool
var queued_requests []int

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
	//id := req.Id

	rep := &ping.Reply{Message: "you're free to go"}
	p.timestamp++
	return rep, nil
}

func (p *peer) sendPingToAll() {
	request := &ping.Request{Id: p.id}
	for id, client := range p.clients {
		reply, err := client.Ping(p.ctx, request)
		if err != nil {
			fmt.Println("something went wrong")
		}
		fmt.Printf("Got reply from id %v: %v\n", id, reply.Message)
		p.timestamp++
	}
}

func priority() {

}

func add_to_request_queue(peer_id int) {
	queued_requests = append(queued_requests, peer_id)
}

/*
func Am_I_priority(pj_id int, lc_pj int) bool {

	if lc_requisicao < lc_pj {

		return true
	} else if lc_pj > lc_requisicao {

		return false
	} else {

		if id < pj_id {

			return true
		} else {

			return false
		}
	}
}


*/
