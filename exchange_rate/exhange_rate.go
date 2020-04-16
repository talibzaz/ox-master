package exchange_rate

import (
	"net/http"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"strings"
	"fmt"
)

const url = "http://data.fixer.io/api/latest?access_key=93738acf7b5a9482422f2ab8c8ef7a0f&base=EUR"

type recievedData struct {
	Rates   map[string]float64 `json:"rates"`
	Success bool               `json:"success"`
}

func GetExchangeRate(symbol string)(float64, error) {

	request, err := http.NewRequest("GET",url, nil)
	if err != nil {
		logrus.Error("error creating request:",err)
		return 0, err
	}
	q := request.URL.Query()
	q.Add("symbols", strings.ToUpper(symbol))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		logrus.Error("error in rquest:",err)
		return 0, err
	}
	defer response.Body.Close()
	fmt.Println(response.Status, response.Header)
	var receivedData recievedData

	if err := json.NewDecoder(response.Body).Decode(&receivedData); err != nil {
		logrus.Error("error decoding response ",err)
		return 0, err
	}
	if receivedData.Success == true {
		return receivedData.Rates[strings.ToUpper(symbol)], nil
	}
	return 0, nil
}