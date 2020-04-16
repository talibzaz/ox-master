package payment_gatway

import (
	s "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	sp "github.com/salfifarooq/shinplasters"
	"github.com/sirupsen/logrus"
	"fmt"
)

/**
 * Created by  ♅ Salfi Farooq on 23/06/18.
 */


func ChargeAmount(currency, stripeToken string, amount uint64) (*s.Charge, error) {

	desc := fmt.Sprintf("Charged %d amount for tickets",amount)
	value, err := sp.GetSubunit(currency)

	if err != nil {
		return nil, err
		logrus.Info("could not get currency subunit", err)
	}
	chargeParam := &s.ChargeParams{
		Amount:   amount * uint64(value),
		Currency: s.Currency(currency),
		Desc:     desc,
	}

	chargeParam.SetSource(stripeToken)
	return charge.New(chargeParam)
}
