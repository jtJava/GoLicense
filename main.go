package main

import (
	"context"
	"github.com/disgoorg/disgo/webhook"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
)

var GlobalWebhookClient webhook.Client

func main() {
	if len(os.Args) < 3 {
		log.Fatal("You didn't set an api key and webhook url.")
	}

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	if err = mongoClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	defer mongoClient.Disconnect(context.TODO())

	GlobalWebhookClient, err = webhook.NewWithURL(os.Args[2])
	if err != nil {
		panic(err)
	}

	collection := mongoClient.Database("default").Collection("licenses")

	handlerContext := HandlerContext{apiKey: os.Args[1], collection: collection}

	router := chi.NewRouter()

	router.Get("/validate/{key}/{hwid}", handlerContext.validateKey)

	// Requires API key
	router.Post("/create", handlerContext.registerKey)
	router.Post("/all", handlerContext.allKeys)
	router.Post("/unused", handlerContext.allUnusedKeys)

	err = http.ListenAndServe(":3000", router)
	if err != nil {
		return
	}
}
