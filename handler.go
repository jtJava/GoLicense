package main

import (
	"Learning/license"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type HandlerContext struct {
	apiKey     string
	collection *mongo.Collection
}

func (ctx *HandlerContext) allUnusedKeys(writer http.ResponseWriter, request *http.Request) {
	ctx.getKeysWithFilter(writer, request, bson.M{"owner": nil})
}

func (ctx *HandlerContext) allKeys(writer http.ResponseWriter, request *http.Request) {
	ctx.getKeysWithFilter(writer, request, nil)
}

func (ctx *HandlerContext) getKeysWithFilter(writer http.ResponseWriter, request *http.Request, filter primitive.M) {
	if !ctx.checkApiKey(request.Header.Get("api-key")) {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var cursor *mongo.Cursor
	var err error
	if filter == nil {
		cursor, err = ctx.collection.Find(context.TODO(), bson.D{})
	} else {
		cursor, err = ctx.collection.Find(context.TODO(), filter)
	}

	if err != nil {
		panic(err)
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	err = json.NewEncoder(writer).Encode(results)
}

func (ctx *HandlerContext) registerKey(writer http.ResponseWriter, request *http.Request) {
	if !ctx.checkApiKey(request.Header.Get("api-key")) {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	licenseKey := license.GenerateKey()
	_, err := ctx.collection.InsertOne(context.TODO(), licenseKey)
	if err != nil {
		return
	}

	err = json.NewEncoder(writer).Encode(licenseKey)
	if err != nil {
		return
	}
}

func (ctx *HandlerContext) validateKey(writer http.ResponseWriter, request *http.Request) {
	key := chi.URLParam(request, "key")
	var licenseKey bson.D
	if err := ctx.collection.FindOne(context.TODO(), bson.M{"_id": key}).Decode(&licenseKey); err != nil {
		_, _ = writer.Write([]byte("false"))
		return
	}
	_, _ = writer.Write([]byte("true"))
}

func (ctx *HandlerContext) checkApiKey(apiKey string) (valid bool) {
	return ctx.apiKey == apiKey
}
