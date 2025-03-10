package flash

import (
	"github.com/gin-gonic/gin"
)

// SetMessage sets a flash message cookie with a default MaxAge of 5 seconds
// This ensures the message is only displayed once and doesn't persist for too long
func SetMessage(ctx *gin.Context, message string, messageType string) {
	SetMessageWithMaxAge(ctx, message, messageType, 5)
}

// SetMessageWithMaxAge sets a flash message cookie with a custom MaxAge in seconds
func SetMessageWithMaxAge(ctx *gin.Context, message string, messageType string, maxAge int) {
	ctx.SetCookie("flash_message", message, maxAge, "/", "", false, true)
	ctx.SetCookie("flash_type", messageType, maxAge, "/", "", false, true)
}

// ClearMessage clears the flash message cookies by setting their MaxAge to -1
func ClearMessage(ctx *gin.Context) {
	ctx.SetCookie("flash_message", "", -1, "/", "", false, true)
	ctx.SetCookie("flash_type", "", -1, "/", "", false, true)
}
