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
	if i_want_to_get_into_citicalsection && req.Timestamp > p.timestamp {
		rep := &ping.Reply{Message: "I was first!"}
		p.timestamp++
		return rep, nil
	} else if i_want_to_get_into_citicalsection && req.Timestamp == p.timestamp && req.Id < p.id {
		rep := &ping.Reply{Message: "I was first!"}
		p.timestamp++
		return rep, nil
	} else {
		rep := &ping.Reply{Message: "you're free to go"}
		p.timestamp++
		return rep, nil
	}
}

func (p *peer) sendPingToAll() {
	request := &ping.Request{Id: p.id, Timestamp: p.timestamp}
	i_want_to_get_into_citicalsection = true
	done := false
	for !done {
		done = true
		for id, client := range p.clients {
			reply, err := client.Ping(p.ctx, request)
			if err != nil {
				fmt.Println("something went wrong")
			}
			if reply.Message == "I was first!" {
				done = false
				break

			} else {
				fmt.Printf("Got reply from id %v: %v\n", id, reply.Message)
				p.timestamp++
			}
		}
	}
	/*Access critical section*/

}

/*func (p *peer) send_priority_to_all() bool {
	//create list of bools
	var responses []bool
	//create waitgroup
	var wg sync.WaitGroup
	for id, client := range p.clients {
		wg.Add(1)
		//create go rutine for each client
		go func(id int32, client ping.PingClient) {
			reply, err := client.Priority(p.ctx, &ping.Request{Id: p.id, Timestamp: p.timestamp})

			if err != nil {
				fmt.Println("something went wrong")
			}

			if p.timestamp > reply.Timestamp {
				responses = append(responses, true)
			} else if reply.Timestamp > p.timestamp {
				responses = append(responses, false)
			} else {
				if p.id < id {
					responses = append(responses, true)
				} else {
					responses = append(responses, false)
				}
			}
			//in go rutine wg.Done
			wg.Done()
		}(id, client)
	}
	wg.Wait()
	//check list if all responses are true
	for _, response := range responses {
		if !response {
			return false
		}
	}
	return true
}

func add_to_request_queue(id int32) {
	queued_requests = append(queued_requests, id)
}*/
