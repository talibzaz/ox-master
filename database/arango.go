package database

import (
	"github.com/spf13/viper"
	"github.com/arangodb/go-driver/http"
	arango "github.com/arangodb/go-driver"
	"context"
)

/**
 * Created by  ♅ Salfi Farooq on 23/06/17.
 */


type ArangoDB struct {
	client arango.Client
	conn   arango.Connection
	ctx    context.Context
}

func NewArangoDB(ctx context.Context) (*ArangoDB, error) {

	conn, err := getConnection()
	if err != nil {
		return nil, err
	}

	client, err := arango.NewClient(arango.ClientConfig{
		Connection: conn,
		Authentication: arango.
			BasicAuthentication(viper.GetString("ARANGO_USERNAME"), viper.GetString("ARANGO_PASSWORD")),
	})

	if err != nil {
		return nil, err
	}

	return &ArangoDB{client: client, conn: conn, ctx: ctx}, nil
}

func (a *ArangoDB) Database(db string) (arango.Database, error) {
	if len(db) == 0 {
		db = viper.GetString("ARANGO_DB")
	}
	return a.client.Database(a.ctx, db)
}

func getConnection() (arango.Connection, error) {
	return http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{viper.GetString("ARANGO_HOST")},
	})
}
