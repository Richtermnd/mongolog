package mongolog

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	URI    string // MongoDB URI
	DBName string // Database name
}

type MongoLogDB struct {
	db *mongo.Database
}

func NewMongoDB(cfg Config) (*MongoLogDB, error) {
	client, err := connect(cfg.URI)
	if err != nil {
		return nil, err
	}
	db := client.Database(cfg.DBName)
	return &MongoLogDB{
		db: db,
	}, nil
}

func (m *MongoLogDB) Insert(ctx context.Context, collection string, data map[string]interface{}) error {
	col := m.db.Collection(collection)
	_, err := col.InsertOne(ctx, data)
	return err
}

func connect(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mongo.Connect(ctx, options.Client().ApplyURI(uri))
}
