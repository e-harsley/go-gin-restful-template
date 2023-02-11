package errorhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotFoundHandler(c *gin.Context, message interface{}) {
	c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": message})
}

func BadRequestHandler(c *gin.Context, message interface{}) {
	c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": message})
}

func ConflictHandler(c *gin.Context, message interface{}) {
	c.JSON(http.StatusConflict, gin.H{"status": http.StatusBadRequest, "error": message})
}

func InternalServerErrorHandler(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error"})
}
