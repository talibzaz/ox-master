package payment_gatway

import (
	"github.com/stripe/stripe-go/token"
	"github.com/stripe/stripe-go"
	"strings"
	"github.com/graphicweave/ox/database/model"
	"github.com/spf13/viper"
	"fmt"
)

/**
 * Created by  ♅ Salfi Farooq on 23/06/17.
 */

func GetToken(card model.Card) (*stripe.Token, error) {
	stripe.Key = viper.GetString("STRIPE_API_KEY")
    fmt.Println(stripe.Key)
	expDate := strings.Split(card.ExpDate, "/");
	return token.New(&stripe.TokenParams{
		Card: &stripe.CardParams{
			Number: card.CardNumber,
			Month:  expDate[0],
			Year:   expDate[1],
			CVC:    card.CVV,
		},
	})

}
