package controllers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/TenJit/SE/Backend/configs"
	"github.com/TenJit/SE/Backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var imageCollection *mongo.Collection = configs.GetCollection(configs.DB, "images")

func generateImagePath(originalFilename string) string {
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(originalFilename)
	return fmt.Sprintf("public/images/%d%s", timestamp, ext)
}

func findLatestExpFolder(outputDir string) (string, error) {
	dirs, err := os.ReadDir(outputDir)
	if err != nil {
		return "", err
	}

	var expDirs []os.DirEntry
	expDirPattern := regexp.MustCompile(`^exp(\d+)$`)

	for _, dir := range dirs {
		if dir.IsDir() && expDirPattern.MatchString(dir.Name()) {
			expDirs = append(expDirs, dir)
		}
	}

	if len(expDirs) == 0 {
		return "", os.ErrNotExist
	}

	sort.Slice(expDirs, func(i, j int) bool {
		iNum, _ := strconv.Atoi(expDirPattern.FindStringSubmatch(expDirs[i].Name())[1])
		jNum, _ := strconv.Atoi(expDirPattern.FindStringSubmatch(expDirs[j].Name())[1])
		return iNum > jNum
	})

	return filepath.Join(outputDir, expDirs[0].Name()), nil
}

func CreateImage(c *gin.Context) {
	user, _ := c.Get("user")
	userData, _ := user.(models.User)

	if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	file, handler, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image from form data"})
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	if name == "" {
		name = "untitled"
	}

	imagePath := generateImagePath(handler.Filename)

	if err := os.MkdirAll(filepath.Dir(imagePath), os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory for image"})
		return
	}

	out, err := os.Create(imagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create image file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image file"})
		return
	}

	image := models.Image{
		ID:                primitive.NewObjectID(),
		ImageName:         name,
		ImagePath:         imagePath,
		User:              userData.ID,
		Status:            "pending",
		CreatedAt:         time.Now(),
		DetectedImagePath: bson.TypeNull.String(),
		Result:            []models.DetectedObject{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = imageCollection.InsertOne(ctx, image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert image into database"})
		return
	}

	baseDir, err := filepath.Abs(".")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get absolute path"})
		return
	}
	detectScriptPath := filepath.Join(baseDir, "imagedetection/yolov5/detect.py")
	weightsPath := filepath.Join(baseDir, "imagedetection/yolov5s-cat-dog.pt")
	outputFolder := filepath.Join(baseDir, "imagedetection/yolov5/runs/detect")
	pythonPath := filepath.Join(baseDir, "imagedetection/yolov5/yolov5_venv/bin/python")

	cmd := exec.Command(pythonPath, detectScriptPath, "--weights", weightsPath, "--img", "640", "--conf", "0.25", "--source", imagePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		_, newerr := imageCollection.UpdateOne(ctx, bson.M{"_id": image.ID}, bson.M{
			"$set": bson.M{
				"status": "fail",
			},
		})
		if newerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update image in database , run detection script", "details": string(output) + "\n" + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run detection script", "details": string(output) + "\n" + err.Error()})
		return
	}

	latestExpFolder, err := findLatestExpFolder(outputFolder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to locate the latest exp folder"})
		return
	}

	jsonFilePath := filepath.Join(latestExpFolder, "detections.json")

	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open JSON output file"})
		return
	}
	defer jsonFile.Close()

	var detectOutput map[string]interface{}
	if err := json.NewDecoder(jsonFile).Decode(&detectOutput); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode JSON output file"})
		return
	}

	results, ok := detectOutput["results"].([]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid JSON structure"})
		return
	}

	detectedObjectsMap := make(map[string][]models.Coordinate)

	for _, result := range results {
		resultMap := result.(map[string]interface{})
		detections := resultMap["detections"].([]interface{})
		for _, detection := range detections {
			detectionMap := detection.(map[string]interface{})
			className := detectionMap["class_name"].(string)
			confidence := float32(detectionMap["confidence"].(float64))
			boundingBox := detectionMap["bounding_box"].(map[string]interface{})

			coordinate := models.Coordinate{
				Bounding_id: primitive.NewObjectID(),
				Confidence:  confidence,
				X_min:       int(boundingBox["x_min"].(float64)),
				X_max:       int(boundingBox["x_max"].(float64)),
				Y_min:       int(boundingBox["y_min"].(float64)),
				Y_max:       int(boundingBox["y_max"].(float64)),
			}
			detectedObjectsMap[className] = append(detectedObjectsMap[className], coordinate)
		}
	}

	var detectedObjects []models.DetectedObject
	for name, coordinates := range detectedObjectsMap {
		detectedObjects = append(detectedObjects, models.DetectedObject{
			Name:        name,
			Coordinates: coordinates,
		})
	}

	detectedImageFiles, err := filepath.Glob(filepath.Join(latestExpFolder, "*.*"))
	if err != nil || len(detectedImageFiles) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find detected image file"})
		return
	}
	detectedImagePath := detectedImageFiles[0]

	detectedImageSavePath := generateImagePath("detected_" + handler.Filename)

	if err := os.MkdirAll(filepath.Dir(detectedImageSavePath), os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory for detected image"})
		return
	}

	srcFile, err := os.Open(detectedImagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open detected image file"})
		return
	}
	defer srcFile.Close()

	destFile, err := os.Create(detectedImageSavePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create detected image file"})
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save detected image file"})
		return
	}

	image.DetectedImagePath = detectedImageSavePath
	image.Result = detectedObjects
	image.Status = "success"

	_, err = imageCollection.UpdateOne(ctx, bson.M{"_id": image.ID}, bson.M{
		"$set": bson.M{
			"result":            image.Result,
			"status":            image.Status,
			"detectedImagePath": image.DetectedImagePath,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update image in database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Image created successfully", "imagePath": image.ImagePath, "image": image})
}

func GetAllImages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user, _ := c.Get("user")
	userData, _ := user.(models.User)

	search := c.Query("search")
	sortBy := c.Query("sortBy")
	status := c.Query("status")
	sortOrder := c.Query("sortOrder")
	recent := c.Query("recent")

	order := -1
	if sortOrder == "asc" {
		order = 1
	}

	matchStage := bson.D{
		{"$match", bson.D{
			{"user", userData.ID},
		}},
	}

	if search != "" {
		matchStage[0].Value = append(matchStage[0].Value.(bson.D), bson.E{"imageName", bson.M{"$regex": search, "$options": "i"}})
	}

	if status != "" {
		matchStage[0].Value = append(matchStage[0].Value.(bson.D), bson.E{"status", status})
	}

	if recent == "true" {
		matchStage[0].Value = append(matchStage[0].Value.(bson.D), bson.E{"createdAt", bson.M{"$gte": time.Now().Add(-24 * time.Hour)}})
	}

	sortField := "createdAt"
	if sortBy == "imageName" {
		sortField = "imageName"
	}

	sortStage := bson.D{
		{"$sort", bson.D{
			{sortField, order},
		}},
	}

	pipeline := mongo.Pipeline{
		matchStage,
		sortStage,
	}

	cursor, err := imageCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var images []models.Image
	if err := cursor.All(ctx, &images); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	counts := len(images)
	c.JSON(http.StatusOK, gin.H{"success": true, "counts": counts, "data": images})
}

func GetImageByID(c *gin.Context) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user, _ := c.Get("user")
	userData, _ := user.(models.User)

	id := c.Param("id")
	ImageID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid image ID"})
		return
	}

	var image models.Image

	err = imageCollection.FindOne(context, bson.M{"_id": ImageID}).Decode(&image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error finding image"})
		return
	}

	if userData.ID == image.User {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": image})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to access this image"})
}

func RenameImage(c *gin.Context) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user, _ := c.Get("user")
	userData, _ := user.(models.User)

	id := c.Param("id")
	ImageID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid image ID"})
		return
	}

	var image models.Image
	err = imageCollection.FindOne(context, bson.M{"_id": ImageID}).Decode(&image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error finding image"})
		return
	}

	if userData.ID != image.User {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to rename this image"})
		return
	}

	newName := c.PostForm("name")
	if newName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Image name is required"})
		return
	}

	update := bson.M{"$set": bson.M{"imageName": newName}}
	_, err = imageCollection.UpdateOne(context, bson.M{"_id": ImageID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error renaming image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Image renamed successfully"})
}

func DeleteImage(c *gin.Context) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user, _ := c.Get("user")
	userData, _ := user.(models.User)

	id := c.Param("id")
	ImageID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid image ID"})
		return
	}

	var image models.Image
	err = imageCollection.FindOne(context, bson.M{"_id": ImageID}).Decode(&image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error finding image"})
		return
	}

	if userData.ID != image.User {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to delete this image"})
		return
	}

	err = os.Remove(image.ImagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error deleting image file"})
		return
	}

	if image.DetectedImagePath != "null" {
		err = os.Remove(image.DetectedImagePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error deleting detected image file"})
			return
		}
	}

	_, err = imageCollection.DeleteOne(context, bson.M{"_id": ImageID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error deleting image from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Image deleted successfully"})
}

func DeleteManyImages(c *gin.Context) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user, _ := c.Get("user")
	userData, _ := user.(models.User)

	var requestData struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	var objectIDs []primitive.ObjectID
	for _, id := range requestData.IDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid image ID"})
			return
		}
		objectIDs = append(objectIDs, objID)
	}

	filter := bson.M{
		"_id":  bson.M{"$in": objectIDs},
		"user": userData.ID,
	}

	var images []models.Image
	cursor, err := imageCollection.Find(context, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error finding images"})
		return
	}
	defer cursor.Close(context)

	if err = cursor.All(context, &images); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error reading images"})
		return
	}

	for _, image := range images {
		err = os.Remove(image.ImagePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Error deleting image file: %s", image.ImagePath)})
			return
		}

		if image.DetectedImagePath != "null" {
			err = os.Remove(image.DetectedImagePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error deleting detected image file"})
				return
			}
		}
	}

	result, err := imageCollection.DeleteMany(context, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error deleting images from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%d images deleted successfully", result.DeletedCount)})
}

func DownloadImage(c *gin.Context) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user, _ := c.Get("user")
	userData, _ := user.(models.User)

	id := c.Param("id")
	ImageID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid image ID"})
		return
	}

	var image models.Image
	err = imageCollection.FindOne(context, bson.M{"_id": ImageID}).Decode(&image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error finding image"})
		return
	}

	if userData.ID != image.User {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Not authorized to download this image"})
		return
	}

	if image.DetectedImagePath != "null" {
		c.FileAttachment(image.DetectedImagePath, "detected_"+image.ImageName)
	} else {
		c.FileAttachment(image.ImagePath, image.ImageName)
	}
}

func DownloadManyImages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user, _ := c.Get("user")
	userData, _ := user.(models.User)

	var requestData struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	var objectIDs []primitive.ObjectID
	for _, id := range requestData.IDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid image ID"})
			return
		}
		objectIDs = append(objectIDs, objID)
	}

	var images []models.Image
	filter := bson.M{"_id": bson.M{"$in": objectIDs}, "user": userData.ID}
	cursor, err := imageCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error finding images"})
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &images); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error reading images"})
		return
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, image := range images {
		file, err := os.Open(image.ImagePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error opening image file"})
			return
		}

		if image.DetectedImagePath != "null" {
			file, err = os.Open(image.DetectedImagePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error opening image file"})
				return
			}
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error getting image file info"})
			return
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error creating zip header"})
			return
		}
		header.Name = image.ImageName + filepath.Ext(image.ImagePath) // Use image name with the correct extension

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error creating zip writer"})
			return
		}

		if _, err := io.Copy(writer, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error writing to zip"})
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error closing zip writer"})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	zipFilename := fmt.Sprintf("images_%s.zip", timestamp)

	c.Writer.Header().Set("Content-Type", "application/zip")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))

	if _, err := io.Copy(c.Writer, buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error writing zip to response"})
		return
	}
}
