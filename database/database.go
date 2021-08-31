package database


import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"context"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
	"gitlab.com/amirkerroumi/my-graphql/model"
)
type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
        log.Fatal("Mongo NewClient error:", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Failure to Connect to MongoDB:", err.Error())
	}

	return &DB {
		client: client,
	}
}

func (db *DB) FindAuthorByID(ID string) *model.Author {
	ObjectID, err := primitive.ObjectIDFromHex(ID)
	collection := db.client.Database("my-gqlgen").Collection("author")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := collection.Find(ctx, bson.M{"_id": ObjectID})
	if err != nil {
		log.Fatal("MongoDB Author Find Failure:", err.Error())
	}
	var author *model.Author
	res.Decode(&author)
	return author
}