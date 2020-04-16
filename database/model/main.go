package model

import (
	"github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
	"context"
	"github.com/graphicweave/ox/database"
)

/**
 * Created by  ♅ Salfi Farooq on 23/06/17.
 */

func GetDatabase(ctx context.Context) (driver.Database, error) {

	arangoDb, err := database.NewArangoDB(ctx)

	if err != nil {
		logrus.Error("couldn't connect to ArangoDB", err)
		return nil, err
	}

	db, err := arangoDb.Database("")
	if err != nil {
		logrus.Error("Purchase Ticket : Error in database ", err)
		return nil, err
	}
	return db, err
}