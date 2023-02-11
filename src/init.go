package src

import (
	"github.com/e-harsley/go-gin-restful-template/src/resources"
	"github.com/e-harsley/go-gin-restful-template/src/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
}

func (app App) RegisterRoutes(router *gin.Engine, resource resources.IResource, basePath string) {
	gin.SetMode(gin.ReleaseMode)
	group := router.Group(basePath)
	group.GET("", resource.Get)
	group.GET("/:id", resource.Get)
	group.POST("/:id/:action", resource.Post)
	group.POST("", resource.Post)
	group.PUT("/:id", resource.Put)
	group.DELETE("", resource.Delete)
}

func GinRestful(db *gorm.DB) App {

	services.FactoryService = services.ServiceFactory{DB: db}

	return App{}

}
