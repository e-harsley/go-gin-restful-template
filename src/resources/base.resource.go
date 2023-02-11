package resources

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type IResource interface {
	Get(ctx *gin.Context)
	Post(ctx *gin.Context)
	Put(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type Resource struct {
}

func (r *Resource) Get(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "This is a GET method for the resource"})
}

func (r *Resource) Post(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "This is a Post method for the resource"})
}

func (r *Resource) Put(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "This is a Put method for the resource"})
}

func (r *Resource) Delete(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "This is a Delete method for the resource"})
}
