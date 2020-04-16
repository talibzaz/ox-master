package model

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"context"
	"github.com/graphicweave/ox/database"
)

/**
 * Created by  ♅ Salfi Farooq on 23/06/17.
 */

type Payment struct {
	CardDetail     Card
	SaveCardDetail bool
	IncludeTax     bool
	ChargeTax      bool
	AmountCharged  float64
	TaxRate        float32
	TicketPrice    float32

}
type Card struct {
	UserID     string
	CardType   string
	CardNumber string
	ExpDate    string
	CVV        string
	Currency   string
}



func (c *Card) SaveCard(userID string) (error) {

	ctx := context.Background()
	c.UserID = userID

	arangoDb, err := database.NewArangoDB(ctx)

	if err != nil {
		logrus.Error("save Card : couldn't connect to ArangoDB", err)
		return err
	}

	db, err := arangoDb.Database("")
	if err != nil {
		logrus.Error("save Card : Error in database ", err)
		return err
	}

	collection, err := db.Collection(ctx, viper.GetString("CARD_COLLECTION"))
	if err != nil {
		logrus.Error("save Card : Error in collection ", err)
		return err
	}

	_, err = collection.CreateDocument(ctx, c)
	if err != nil {
		logrus.Error("save Card : Couldn't create document", err)
		return err
	}

	return nil

}

