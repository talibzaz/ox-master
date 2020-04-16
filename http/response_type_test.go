package http

import (
	"testing"
	"github.com/spf13/viper"
	"encoding/json"
	"os"
)

func TestTicketDetail_GetTicketDetails(t *testing.T) {

	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	detail := TicketDetail{}

	details := detail.GetTicketDetails("bdru546p12qt4gubere0")

	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(details)
}
