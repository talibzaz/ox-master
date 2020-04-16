package exchange_rate

import (
	"testing"
	"fmt"
)

func TestGetExchangeRate(t *testing.T) {
	recievedData, err := GetExchangeRate("INR")
	if err != nil {
		t.Error("error getting exchange rate")
	}
	fmt.Println(recievedData)
}