package license

import (
	"github.com/google/uuid"
	"time"
)

func GenerateKey() License {
	return License{Key: uuid.New().String(), CreationDate: time.Now()}
}

type License struct {
	Key          string    `json:"id" bson:"_id"`
	CreationDate time.Time `json:"creation_date" bson:"creation_date"`
	Disabled     bool      `json:"disabled" bson:"disabled"`
}
