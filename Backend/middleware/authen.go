package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/TenJit/SE/Backend/configs"
	"github.com/TenJit/SE/Backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")

func Protect(c *gin.Context) {
	tokenString, err := extractTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to access this route"})
		c.Abort()
		return
	}

	claims, err := verifyToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to access this route"})
		c.Abort()
		return
	}

	userID, ok := claims["id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to access this route"})
		c.Abort()
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to access this route"})
		c.Abort()
		return
	}

	var user models.User
	if err := userCollection.FindOne(c, bson.M{"_id": objectID}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to access this route"})
		c.Abort()
		return
	}

	c.Set("user", user)
	c.Next()
}

func Authorize(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to access this route"})
			c.Abort()
			return
		}

		userModel, ok := user.(models.User)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to access this route"})
			c.Abort()
			return
		}

		for _, role := range roles {
			if userModel.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "User role is not authorized to access this route"})
		c.Abort()
	}
}

func extractTokenFromHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header provided")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

func verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(configs.JWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
