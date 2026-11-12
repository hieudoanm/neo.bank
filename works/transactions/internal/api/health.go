package api

import (
	"context"
	"net/http"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/gin-gonic/gin"
)

func HealthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func ReadyzHandler(session *gocql.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		if session == nil {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		var version string
		if err := session.Query("SELECT release_version FROM system.local").WithContext(ctx).Scan(&version); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
