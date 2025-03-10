package flash

import (
	"github.com/gin-gonic/gin"
)

// SetMessage sets a flash message cookie with a default MaxAge of 3600 seconds (1 hour)
func SetMessage(ctx *gin.Context, message string, messageType string) {
	SetMessageWithMaxAge(ctx, message, messageType, 3600)
}

// SetMessageWithMaxAge sets a flash message cookie with a custom MaxAge in seconds
func SetMessageWithMaxAge(ctx *gin.Context, message string, messageType string, maxAge int) {
	ctx.SetCookie("flash_message", message, maxAge, "/", "", false, true)
	ctx.SetCookie("flash_type", messageType, maxAge, "/", "", false, true)
}
