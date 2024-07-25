package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/TenJit/SE/Backend/configs"
	"github.com/TenJit/SE/Backend/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")
var validateUser = validator.New()

func createToken(id primitive.ObjectID) (string, error) {
	claims := jwt.MapClaims{}
	claims["id"] = id
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token valid for 1 hour

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := []byte(configs.JWTSecret())
	println(jwtSecret)
	return token.SignedString(jwtSecret)
}

func Register(c *gin.Context) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var user models.User
	defer cancel()
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input", "error": err.Error()})
		return
	}

	if validationErr := validateUser.Struct(&user); validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input", "error": validationErr.Error()})
		return
	}

	validTel := IsValidPhoneNumber(user.Tel) || user.Tel == ""
	validEmail := IsValidEmail(user.Email)
	if !validTel || !validEmail {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "valid email": validEmail, "valid tel": validTel})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error hashing password", "error": err.Error()})
		return
	}

	newUser := models.User{
		ID:        primitive.NewObjectID(),
		Name:      user.Name,
		Password:  string(hashedPassword),
		Email:     user.Email,
		Tel:       user.Tel,
		Role:      user.Role,
		CreatedAt: time.Now(),
	}

	result, err := userCollection.InsertOne(context, newUser)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error registering", "error": err.Error()})
		return
	}

	insertedID, _ := result.InsertedID.(primitive.ObjectID)

	token, _ := createToken(insertedID)

	c.SetCookie("token", token, configs.JWTCookieExpire()*24*60*60, "/", "", false, true)

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Created User Successfully", "_id": insertedID, "token": token})
}

func LogIn(c *gin.Context) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var credential models.UserLogIn
	defer cancel()
	if err := c.ShouldBindJSON(&credential); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid credentials", "error": err.Error()})
		return
	}

	if validationErr := validateUser.Struct(&credential); validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid credentials", "error": validationErr.Error()})
		return
	}

	var user models.User
	if err := userCollection.FindOne(context, bson.M{"email": credential.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid credentials", "error": err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credential.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid credentials", "error": err.Error()})
		return
	}

	token, _ := createToken(user.ID)
	c.SetCookie("token", token, configs.JWTCookieExpire()*24*60*60, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Log in successfully", "token": token})
}

func GetMe(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Not logged in"})
		return
	}

	userData, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to retrieve user data"})
		return
	}

	userResponse := models.UserResponse{
		ID:        userData.ID,
		Name:      userData.Name,
		Email:     userData.Email,
		Tel:       userData.Tel,
		Role:      userData.Role,
		CreatedAt: userData.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": userResponse})
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Not logged in"})
		return
	}

	userData, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to retrieve user data"})
		return
	}

	if userData.Role != "admin" {
		if userData.ID != userID {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "You don't have access to this user id"})
			return
		}
	}

	var updateReq models.UserUpdate
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid inputtt", "error": err.Error()})
		return
	}

	if validationErr := validateUser.Struct(&updateReq); validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input", "error": validationErr.Error()})
		return
	}

	if !IsValidPhoneNumber(updateReq.Tel) && updateReq.Tel != "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid telephone number"})
		return
	}

	result := userCollection.FindOneAndUpdate(context.Background(), bson.M{"_id": userID}, bson.M{"$set": updateReq})
	if result.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error updating user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User updated successfully"})
}

func LogOut(c *gin.Context) {
	c.SetCookie("token", "", 10*1000, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Log out successfully"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Not logged in"})
		return
	}

	userData, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to retrieve user data"})
		return
	}

	if userData.Role != "admin" {
		if userData.ID != userID {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "You don't have access to this user id"})
			return
		}
	}

	result := userCollection.FindOneAndDelete(context.TODO(), bson.M{"_id": userID})
	if result.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error deleting user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted successfully"})
}

func GetAllUser(c *gin.Context) {
	projection := bson.M{"password": 0, "resetPasswordToken": 0, "resetPasswordExpire": 0}

	cursor, err := userCollection.Find(context.TODO(), bson.D{{}}, options.Find().SetProjection(projection))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var user []models.UserResponse
	if err = cursor.All(context.TODO(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	counts := len(user)
	c.JSON(http.StatusOK, gin.H{"success": true, "count": counts, "data": user})
}
