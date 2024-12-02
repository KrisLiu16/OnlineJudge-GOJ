package middleware

import (
	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 不记录日志，直接执行下一个中间件
		c.Next()
	}
}
