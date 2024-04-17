package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type CollSize struct {
	MaxSize int64   `bson:"maxSize"` // Shows the maximum size of the collection.
	Size    float64 `bson:"size"`    // The total size in memory of all records in a collection.
}

type OpTime struct {
	Ts primitive.Timestamp `bson:"ts"` // The timestamp of the last operation applied to this member of the replica set
	T  int                 `bson:"t"`  // The term in which the last applied operation was originally generated on the primary.
}
