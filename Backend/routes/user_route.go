package routes

import (
	"github.com/TenJit/SE/Backend/controllers"
	"github.com/TenJit/SE/Backend/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoute(app *gin.Engine) {
	app.POST("/register", controllers.Register)
	app.POST("/login", controllers.LogIn)
	app.GET("/getme", middleware.Protect, controllers.GetMe)
	app.GET("/logout", controllers.LogOut)
	app.GET("/users", middleware.Protect, middleware.Authorize("admin"), controllers.GetAllUser)
	app.PUT("/updateuser/:id", middleware.Protect, controllers.UpdateUser)
	app.DELETE("/deleteuser/:id", middleware.Protect, controllers.DeleteUser)
}
