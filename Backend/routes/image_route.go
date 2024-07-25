package routes

import (
	"github.com/TenJit/SE/Backend/controllers"
	"github.com/TenJit/SE/Backend/middleware"

	"github.com/gin-gonic/gin"
)

func ImageRoute(app *gin.Engine) {
	imagesRoutes := app.Group("/images")
	{
		protectedRoutes := imagesRoutes.Group("", middleware.Protect)
		{
			protectedRoutes.POST("", controllers.CreateImage)
			protectedRoutes.GET("", controllers.GetAllImages)
			protectedRoutes.GET("/:id", controllers.GetImageByID)
			protectedRoutes.PUT("/:id", controllers.RenameImage)
			protectedRoutes.DELETE("/:id", controllers.DeleteImage)
			protectedRoutes.DELETE("", controllers.DeleteManyImages)
			protectedRoutes.GET("/download/:id", controllers.DownloadImage)
			protectedRoutes.POST("/downloadManyImages", controllers.DownloadManyImages)
		}
	}
}
