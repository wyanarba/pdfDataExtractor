package models

import "go.mongodb.org/mongo-driver/v2/bson"

type DailyPage struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	Day      int           `bson:"day"`
	Hash     string        `bson:"hash"`
	Page     int           `bson:"page"`
	Building int           `bson:"building"`
}
