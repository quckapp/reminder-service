package service

import "go.mongodb.org/mongo-driver/bson/primitive"

func objectIDFromHex(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}
