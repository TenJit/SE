package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Image struct {
	ID                primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	User              primitive.ObjectID `json:"user,omitempty" bsom:"user,omitempty"`
	ImageName         string             `json:"imageName,omitempty" bson:"imageName,omitempty"`
	ImagePath         string             `json:"imagePath,omitempty" bson:"imagePath,omitempty" validate:"required"`
	DetectedImagePath string             `json:"detectedImagePath,omitempty" bson:"detectedImagePath,omitempty"`
	Status            string             `json:"status,omitempty" bson:"status,omitempty"`
	Result            []DetectedObject   `json:"result" bson:"result"`
	CreatedAt         time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type DetectedObject struct {
	Name        string       `json:"name,omitempty" bson:"name,omitempty"`
	Coordinates []Coordinate `json:"coordinates,omitempty" bson:"coordinates,omitempty"`
}

type Coordinate struct {
	Bounding_id primitive.ObjectID `json:"bounding_id,omitempty" bson:"bounding_id,omitempty"`
	Confidence  float32            `json:"confidence,omitempty" bson:"confidence,omitempty"`
	X_min       int                `json:"x_min,omitempty" bson:"x_min,omitempty"`
	X_max       int                `json:"x_max,omitempty" bson:"x_max,omitempty"`
	Y_min       int                `json:"y_min,omitempty" bson:"y_min,omitempty"`
	Y_max       int                `json:"y_max,omitempty" bson:"y_max,omitempty"`
}
