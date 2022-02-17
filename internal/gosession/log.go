package gosession

import "go.mongodb.org/mongo-driver/bson/primitive"

/*
	LOG SYSTEM

	These logs should be saved into the s3 bucket
*/

// UserLog struct
type UserLog struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Title string             `bson:"title,omitempty"`
}
