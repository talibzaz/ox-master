package event

import (
	"testing"
	"google.golang.org/grpc"
	"github.com/graphicweave/ox/ox_idl"
	"io"
	"log"
	"context"
	"fmt"
)

/**
* Created by  ♅ Salfi Farooq on 23/06/17.
*/

func TestOXService_GetTicketsByID(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	ox := ox_idl.NewOXClient(conn)
	d := &ox_idl.TicketRequest{Id: "5770822d-f8e8-4f90-9242-73dea0efe409"}
	fmt.Println(d)
	s, err := ox.GetTicketsByID(context.Background(), d)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("asdasd",s)

	for {
		feature, err := s.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		log.Println(feature)
	}
	defer conn.Close()
}

func TestOXService_GetAttendeesStatus(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	ox := ox_idl.NewOXClient(conn)
	d := &ox_idl.TicketRequest{Id: "bcf3dv1ruqip8rgggjqg"}
	s, err := ox.GetAttendeesStatus(context.Background(),d)
	if err != nil {
	   t.Fatal(err)
		return
	}
	fmt.Println(s)

}