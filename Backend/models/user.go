package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name                string             `json:"name,omitempty" bson:"name,omitempty" validate:"required"`
	Tel                 string             `json:"tel,omitempty" bson:"tel,omitempty"`
	Email               string             `json:"email,omitempty" bson:"email,omitempty" validate:"required"`
	Role                string             `json:"role,omitempty" bson:"role,omitempty" validate:"required"`
	Password            string             `json:"password,omitempty" bson:"password,omitempty" validate:"required"`
	ResetPasswordToken  string             `json:"resetPasswordToken,omitempty" bson:"resetPasswordToken,omitempty"`
	ResetPasswordExpire time.Time          `json:"resetPasswordExpire,omitempty" bson:"resetPasswordExpire,omitempty"`
	CreatedAt           time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type UserLogIn struct {
	Email    string `json:"email,omitempty" bson:"email,omitempty" validate:"required"`
	Password string `json:"password,omitempty" bson:"password,omitempty" validate:"required"`
}

type UserResponse struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name,omitempty" validate:"required"`
	Tel       string             `json:"tel,omitempty" bson:"tel,omitempty"`
	Email     string             `json:"email,omitempty" bson:"email,omitempty" validate:"required"`
	Role      string             `json:"role,omitempty" bson:"role,omitempty" validate:"required"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type UserUpdate struct {
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	Tel  string `json:"tel,omitempty" bson:"tel,omitempty"`
}
