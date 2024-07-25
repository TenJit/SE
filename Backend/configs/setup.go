package configs

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect_to_mongodb() *mongo.Client {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(EnvMongoURI()).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	return client
}

// Client instance
var DB *mongo.Client = Connect_to_mongodb()

// getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("ImageDetection").Collection(collectionName)
	return collection
}
