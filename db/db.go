package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var conn *mongo.Client

func initDb() {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	conn = client
	if err != nil {
		log.Fatal(err)
	}
}

func GetDb() *mongo.Client {
	if conn == nil {
		initDb()
	}

	return conn
}

func CloseDb() {
	conn.Disconnect(context.TODO())
}
