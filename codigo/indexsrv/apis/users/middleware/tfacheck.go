package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/dtos/jsend"
	"github.com/mredolatti/tf/codigo/common/log"
)

type TFACheck struct {
	logger log.Interface
}

func NewTFACheck(logger log.Interface) *TFACheck {
	return &TFACheck{ logger: logger, }
}

func (t *TFACheck) Handle(ctx *gin.Context) {
	session, err := SessionFromContext(ctx)
	if err != nil { // we should have a session by now
		t.logger.Error("error getting session from context: %s", err)
		ctx.AbortWithStatus(500)
		return
	}

	if !session.TFADone() {
//		ctx.AbortWithStatusJSON(403, gin.H{"message": "2fa is required to access this endpoint"})
		ctx.AbortWithStatusJSON(403, jsend.NewCustomFailResponse("2fa is required to access this endpoint", "X-IS-MIFS-Token", "missing OTP"))
		return
	}
}
