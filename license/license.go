package license

import (
	"github.com/google/uuid"
	"time"
)

func GenerateKey(owner string, project string) License {
	return License{Key: uuid.New().String(), Owner: owner, Project: project, CreationDate: time.Now()}
}

func GenerateKeyWithHWID(owner string, project string, hwid string) License {
	return License{Key: uuid.New().String(), Owner: owner, Project: project, HWIDRequired: true, HWID: hwid, CreationDate: time.Now()}
}

type License struct {
	Key          string    `json:"id" bson:"_id"`
	Owner        string    `json:"owner" bson:"owner"`
	Project      string    `json:"project" bson:"project"`
	HWID         string    `json:"hwid" bson:"hwid"`
	HWIDRequired bool      `json:"hwid_required" bson:"hwid_required"`
	CreationDate time.Time `json:"creation_date" bson:"creation_date"`
	Disabled     bool      `json:"disabled" bson:"disabled"`
}
