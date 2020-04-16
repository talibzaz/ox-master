package model

import (
	"testing"
	"github.com/spf13/viper"
	"encoding/json"
	"os"
	"fmt"
)

func TestGetAttendees(t *testing.T) {

	viper.Set("ARANGO_HOST", "http://localhost:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "QAUSohCH9wxiA4KW")

	details, err := GetAttendees("bcf3dv1ruqip8rgggjqg", true)

	if err != nil {
		t.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", " ")
	enc.Encode(&details)
}

func TestGetTicketsByUserID(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")


}
func TestCheckAvailability(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")
	viper.Set("EVENTS_COLLECTION", "events")

	e, er := CheckAvailability("bd2p5q70qrog00f20280",2)
	if er != nil {
		t.Fatal(er)
	}
	fmt.Println(e)
}
func TestAttendeeCount(t *testing.T) {

	viper.Set("ARANGO_HOST", "http://localhost:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "QAUSohCH9wxiA4KW")

	c,e := AttendeeCount("bcf3dv1ruqip8rgggjqg")
	fmt.Println(c.AttendeeCount,e)
}

func TestGetTicketDetailsById(t *testing.T) {

	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	details, err := GetTicketDetailsById("beqqh8nhnklp3jv1qolg")
	if err != nil {
		t.Fatal(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(details)

}

func TestPrintTicketForAttendee(t *testing.T) {

	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	resp, err := PrintTicketForAttendee("bdru546p12qt4gubere0", "bdru566p12qt4gubereg")

	if err != nil {
		t.Fatal(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(resp)

}


func TestPrintAllAttendeeTickets(t *testing.T) {

	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	resp, err := PrintAllAttendeeTickets("bdru546p12qt4gubere0")

	if err != nil {
		t.Fatal(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(resp)

}

func TestGetAttendeeInfo(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	resp, err := GetAttendeeInfo("bepi3p6p12qtsamtvtkg")

	if err != nil {
		t.Fatal(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(resp)

}

func TestGetBillingDetail(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")
	viper.Set("CARD_COLLECTION", "card_details")
	resp, err := GetBillingDetail("6393ffc5-9ba0-429d-assssaa3-f938857afdbb")

	if err != nil {
		t.Fatal("bf",err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(resp)
}