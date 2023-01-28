package main

import (
	"Learning/license"
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/discord"
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

	var licenseKey license.License
	owner := request.Header.Get("owner")
	project := request.Header.Get("project")
	hwid := request.Header.Get("hwid")
	if hwid != "" {
		licenseKey = license.GenerateKeyWithHWID(owner, project, request.Header.Get("hwid"))
	} else {
		licenseKey = license.GenerateKey(owner, project)
	}

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
	hwid := chi.URLParam(request, "hwid")
	var licenseKey bson.D
	if err := ctx.collection.FindOne(context.TODO(), bson.M{"_id": key}).Decode(&licenseKey); err != nil {
		_, _ = writer.Write([]byte("false"))
		embed := []discord.Embed{discord.NewEmbedBuilder().
			SetColor(16711680).
			SetTitle("Suspicious License Request").
			AddField("License Key", key, true).
			AddField("IP Address", request.RemoteAddr, false).
			Build()}
		_, _ = GlobalWebhookClient.CreateEmbeds(embed)
		return
	}

	var retrievedLicense license.License
	bsonData, err := bson.Marshal(licenseKey)

	if err != nil {
		fmt.Println("There was an issue marshalling license key: ", key)
		_, _ = writer.Write([]byte("error 1"))
		return
	}

	if err = bson.Unmarshal(bsonData, &retrievedLicense); err != nil {
		fmt.Println("There was an issue unmarshalling license key: ", key)
		_, _ = writer.Write([]byte("error 2"))
		return
	}

	_, _ = writer.Write([]byte("true"))

	if retrievedLicense.HWIDRequired {
		if retrievedLicense.HWID != hwid {
			_, _ = writer.Write([]byte("false"))
			embed := []discord.Embed{discord.NewEmbedBuilder().
				SetColor(16711680).
				SetTitle("Suspicious HWID License Request").
				AddField("License Key", key, true).
				AddField("Owner", retrievedLicense.Owner, true).
				AddField("IP Address", request.RemoteAddr, false).
				SetFooterText("Required HWID: " + retrievedLicense.HWID + " Received HWID: " + hwid).
				Build()}
			_, _ = GlobalWebhookClient.CreateEmbeds(embed)
			return
		}
	}

	if retrievedLicense.Disabled {
		embed := []discord.Embed{discord.NewEmbedBuilder().
			SetColor(16711680).
			SetTitle("Disabled License Request").
			AddField("License Key", key, true).
			AddField("Owner", retrievedLicense.Owner, true).
			AddField("IP Address", request.RemoteAddr, false).
			SetFooterText("Required HWID: " + retrievedLicense.HWID + " Received HWID: " + hwid).
			Build()}
		_, _ = GlobalWebhookClient.CreateEmbeds(embed)
	}

	embed := []discord.Embed{discord.NewEmbedBuilder().
		SetColor(65280).
		SetTitle("License Verified Successfully").
		AddField("License Key", key, false).
		AddField("Owner", retrievedLicense.Owner, true).
		AddField("IP Address", request.RemoteAddr, true).
		AddField("Creation Date", retrievedLicense.CreationDate.String(), false).
		Build()}
	_, _ = GlobalWebhookClient.CreateEmbeds(embed)
}

func (ctx *HandlerContext) checkApiKey(apiKey string) (valid bool) {
	return ctx.apiKey == apiKey
}
