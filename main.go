package main

import (
	"Learning/license"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
)

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	collection = client.Database("default").Collection("licenses")

	router := chi.NewRouter()
	router.Get("/create/{key}", registerKey)
	router.Get("/validate/{key}", validateKey)

	err = http.ListenAndServe(":3000", router)
	if err != nil {
		return
	}
}

var collection *mongo.Collection

var licenses = make(map[string]license.License)

func registerKey(writer http.ResponseWriter, request *http.Request) {
	licenseKey := license.GenerateKey()
	_, err := collection.InsertOne(context.TODO(), licenseKey)
	if err != nil {
		return
	}

	licenses[licenseKey.Key] = licenseKey
	err = json.NewEncoder(writer).Encode(licenseKey)
	if err != nil {
		return
	}
}

func validateKey(writer http.ResponseWriter, request *http.Request) {
	key := chi.URLParam(request, "key")
	var result bson.D
	err := collection.FindOne(context.TODO(), bson.M{"_id": key}).Decode(&result)
	var podcast bson.D
	if err = collection.FindOne(context.TODO(), bson.M{"_id": key}).Decode(&podcast); err != nil {
		log.Fatal(err)
	}
	doc, err := bson.Marshal(bson.M{"_id": key})
	fmt.Println(podcast)
	var licensekey license.License
	err = bson.Unmarshal(doc, &licensekey)
	if err != nil {
		println(licensekey.Disabled)
		return
	}

	println(&result)
	//for _, licens := range licenses {
	//	if licens.Key.String() != key {
	//		continue
	//	}
	//
	//	err := json.NewEncoder(writer).Encode(licens)
	//	if err != nil {
	//		println("Match found for " + key)
	//		return
	//	}
	//}
}
