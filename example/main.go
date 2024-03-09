package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"

	"github.com/Richtermnd/mongolog"
)

type Mock struct {
}

func (*Mock) Insert(ctx context.Context, collection string, data map[string]interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Way to connect to mongo
func ConnectToMongo() mongolog.Inserter {
	db, err := mongolog.NewMongoDB(mongolog.Config{
		URI:    "mongodb://root:password@localhost:27017",
		DBName: "logs",
	})
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {
	// Inserter mock
	inserter := &Mock{}

	// For different apps you can use different mongodb collections.
	handler := mongolog.NewMongoHandler(inserter, "app", slog.LevelDebug)
	log := slog.New(handler)
	log.Info("test1", slog.String("key", "value"))

	log = log.With(slog.String("attr", "value"))
	log.Info("test2", slog.String("key", "value"))

	log = log.WithGroup("group1").With(slog.String("attr1", "value1")).WithGroup("group2").With(slog.String("attr2", "value2"))
	log.Info("test3", slog.String("key", "value"))
}
