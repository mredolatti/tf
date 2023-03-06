package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/access/authentication"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
)

const (
	sessionAuthHeaderName = "X-MIFS-IS-Session-Token"
	sessionCtxKey = "CTX_S_KEY"
	sessionTokenCtxKey = "CTX_S_TOKEN"
)

var (
	ErrNoSessionData = errors.New("no session data in request context")
	ErrInvalidSessionData = errors.New("invalid session data in request context")
)

type SessionAuth struct {
	sessionStore authentication.SessionManager
	logger       log.Interface
}

func NewSessionAuth(sessionStore authentication.SessionManager, logger log.Interface) *SessionAuth {
	return &SessionAuth{
		sessionStore: sessionStore,
		logger: logger,
	}
}

func (m *SessionAuth) Handle(ctx *gin.Context) {

	sessionID := ctx.Request.Header.Get(sessionAuthHeaderName)
	if sessionID == "" {
		m.logger.Error("Invalid request: session token missing in headers.")
		ctx.AbortWithStatus(400)
		return
	}

	session, err := m.sessionStore.LookUp(ctx.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, authentication.ErrNoSuchSession) {
			m.logger.Error("Token '%s' provided in request not found", sessionID)
			ctx.AbortWithStatus(401)
		} else {
			m.logger.Error("error fetching session for id '%s': %s", err)
			ctx.AbortWithStatus(500)
		}
		return
	}

	ctx.Set(sessionCtxKey, session)
	ctx.Set(sessionTokenCtxKey, sessionID)
}

func SessionFromContext(ctx *gin.Context) (models.Session, error) {
	val, exists := ctx.Get(sessionCtxKey)
	if !exists {
		return nil, ErrNoSessionData
	}

	asSession, ok := val.(models.Session)
	if !ok {
		return nil, ErrInvalidSessionData
	}

	return asSession, nil
}

func SessionTokenFromContext(ctx *gin.Context) (string ,error) {
	val, exists := ctx.Get(sessionTokenCtxKey)
	if !exists {
		return "", ErrNoSessionData
	}

	asString, ok := val.(string)
	if !ok {
		return "", ErrInvalidSessionData
	}

	return asString, nil
}
