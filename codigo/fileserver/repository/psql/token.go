package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/mredolatti/tf/codigo/fileserver/models"
	"github.com/mredolatti/tf/codigo/fileserver/repository"

	"github.com/jmoiron/sqlx"
)

const (
	tokenCreate          = "INSERT INTO tokens(client_id, user_id, redirect_uri, scope, code, code_created_at, code_expires_in_seconds, code_challenge, code_challenge_method, access, access_created_at, access_expires_in_seconds, refresh, refresh_created_at, refresh_expires_in_seconds) VALUES(:client_id, :user_id, :redirect_uri, :scope, :code, :code_created_at, :code_expires_in_seconds, :code_challenge, :code_challenge_method, :access, :access_created_at, :access_expires_in_seconds, :refresh, :refresh_created_at, :refresh_expires_in_seconds)"
	tokenGetByCode       = "SELECT * FROM tokens WHERE code = $1"
	tokenGetByAccess     = "SELECT * FROM tokens WHERE access = $1"
	tokenGetByRefresh    = "SELECT * FROM tokens WHERE refresh = $1"
	tokenRemoveByCode    = "DELETE FROM tokens WHERE code = $1"
	tokenRemoveByAccess  = "DELETE FROM tokens WHERE access = $1"
	tokenRemoveByRefresh = "DELETE FROM tokens WHERE refresh = $1"
)

// TokenInfo is a postgres-compatible struct implementing models.TokenInfo interface
type TokenInfo struct {
	ClientIDField            string    `db:"client_id"`
	UserIDField              string    `db:"user_id"`
	RedirectURIField         string    `db:"redirect_uri"`
	ScopeField               string    `db:"scope"`
	CodeField                string    `db:"code"`
	CodeCreateAtField        time.Time `db:"code_created_at"`
	CodeExpiresInField       int64     `db:"code_expires_in_seconds"`
	CodeChallengeField       string    `db:"code_challenge"`
	CodeChallengeMethodField string    `db:"code_challenge_method"`
	AccessField              string    `db:"access"`
	AccessCreateAtField      time.Time `db:"access_created_at"`
	AccessExpiresInField     int64     `db:"access_expires_in_seconds"`
	RefreshField             string    `db:"refresh"`
	RefreshCreateAtField     time.Time `db:"refresh_created_at"`
	RefreshExpiresInField    int64     `db:"refresh_expires_in_seconds"`
}

func (t *TokenInfo) New() models.TokenInfo {
	return &TokenInfo{}
}

func (t *TokenInfo) GetClientID() string {
	return t.ClientIDField
}

func (t *TokenInfo) GetUserID() string {
	return t.UserIDField
}

func (t *TokenInfo) GetRedirectURI() string {
	return t.RedirectURIField
}

func (t *TokenInfo) GetScope() string {
	return t.ScopeField
}

func (t *TokenInfo) GetCode() string {
	return t.CodeField
}

func (t *TokenInfo) GetCodeCreateAt() time.Time {
	return t.CodeCreateAtField
}

func (t *TokenInfo) GetCodeExpiresIn() time.Duration {
	return time.Duration(t.CodeExpiresInField) * time.Second
}

func (t *TokenInfo) GetCodeChallenge() string {
	return t.CodeChallengeField
}

func (t *TokenInfo) GetCodeChallengeMethod() oauth2.CodeChallengeMethod {
	return oauth2.CodeChallengeMethod(t.CodeChallengeMethodField)
}

func (t *TokenInfo) GetAccess() string {
	return t.AccessField
}

func (t *TokenInfo) GetAccessCreateAt() time.Time {
	return t.AccessCreateAtField
}

func (t *TokenInfo) GetAccessExpiresIn() time.Duration {
	return time.Duration(t.AccessExpiresInField) * time.Second
}

func (t *TokenInfo) GetRefresh() string {
	return t.RefreshField
}

func (t *TokenInfo) GetRefreshCreateAt() time.Time {
	return t.RefreshCreateAtField
}

func (t *TokenInfo) GetRefreshExpiresIn() time.Duration {
	return time.Duration(t.RefreshExpiresInField) * time.Second
}

func (t *TokenInfo) SetClientID(clientID string) {
	t.ClientIDField = clientID
}

func (t *TokenInfo) SetUserID(userID string) {
	t.UserIDField = userID
}

func (t *TokenInfo) SetRedirectURI(redirectURI string) {
	t.RedirectURIField = redirectURI
}

func (t *TokenInfo) SetScope(scope string) {
	t.ScopeField = scope
}

func (t *TokenInfo) SetCode(code string) {
	t.CodeField = code
}

func (t *TokenInfo) SetCodeCreateAt(codeCreateAt time.Time) {
	t.CodeCreateAtField = codeCreateAt
}

func (t *TokenInfo) SetCodeExpiresIn(codeExpiresIn time.Duration) {
	t.CodeExpiresInField = int64(codeExpiresIn.Seconds())
}

func (t *TokenInfo) SetCodeChallenge(codeChallenge string) {
	t.CodeChallengeField = codeChallenge
}

func (t *TokenInfo) SetCodeChallengeMethod(codeChallengeMethod oauth2.CodeChallengeMethod) {
	t.CodeChallengeMethodField = string(codeChallengeMethod)
}

func (t *TokenInfo) SetAccess(access string) {
	t.AccessField = access
}

func (t *TokenInfo) SetAccessCreateAt(accessCreateAt time.Time) {
	t.AccessCreateAtField = accessCreateAt
}

func (t *TokenInfo) SetAccessExpiresIn(accessExpiresIn time.Duration) {
	t.AccessExpiresInField = int64(accessExpiresIn.Seconds())
}

func (t *TokenInfo) SetRefresh(refresh string) {
	t.RefreshField = refresh
}

func (t *TokenInfo) SetRefreshCreateAt(refreshCreateAt time.Time) {
	t.RefreshCreateAtField = refreshCreateAt
}

func (t *TokenInfo) SetRefreshExpiresIn(refreshExpiresIn time.Duration) {
	t.RefreshExpiresInField = int64(refreshExpiresIn.Seconds())
}

// TokenInfoRepository is a mapping to a table in postgres that allows enables operations
// on file server
type TokenInfoRepository struct {
	db *sqlx.DB
}

func NewTokenInfoRepository(db *sqlx.DB) (*TokenInfoRepository, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	return &TokenInfoRepository{db: db}, nil
}

func (r *TokenInfoRepository) Create(ctx context.Context, info models.TokenInfo) error {
	var tokenForDB *TokenInfo
	var ok bool
	if tokenForDB, ok = info.(*TokenInfo); !ok {
		tokenForDB = &TokenInfo{
			ClientIDField:            info.GetClientID(),
			UserIDField:              info.GetUserID(),
			RedirectURIField:         info.GetRedirectURI(),
			ScopeField:               info.GetScope(),
			CodeField:                info.GetCode(),
			CodeCreateAtField:        info.GetCodeCreateAt(),
			CodeExpiresInField:       int64(info.GetCodeExpiresIn().Seconds()),
			CodeChallengeField:       info.GetCodeChallenge(),
			CodeChallengeMethodField: string(info.GetCodeChallengeMethod()),
			AccessField:              info.GetAccess(),
			AccessCreateAtField:      info.GetAccessCreateAt(),
			AccessExpiresInField:     int64(info.GetAccessExpiresIn().Seconds()),
			RefreshField:             info.GetRefresh(),
			RefreshCreateAtField:     info.GetRefreshCreateAt(),
			RefreshExpiresInField:    int64(info.GetRefreshExpiresIn().Seconds()),
		}
		// TODO(mredolatti): Convert!
	}
	if _, err := r.db.NamedExecContext(ctx, tokenCreate, tokenForDB); err != nil {
		return fmt.Errorf("error executing token_repository::create in postgres: %w", err)
	}

	return nil
}

func (r *TokenInfoRepository) GetByCode(ctx context.Context, code string) (models.TokenInfo, error) {
	var token TokenInfo
	if err := r.db.QueryRowxContext(ctx, tokenGetByCode, code).StructScan(&token); err != nil {
		return nil, fmt.Errorf("error executing token_repository::get_by_code in postgres: %w", err)
	}

	return &token, nil
}

func (r *TokenInfoRepository) GetByAccess(ctx context.Context, access string) (models.TokenInfo, error) {
	var token TokenInfo
	if err := r.db.QueryRowxContext(ctx, tokenGetByCode, access).StructScan(&token); err != nil {
		return nil, fmt.Errorf("error executing token_repository::get_by_access in postgres: %w", err)
	}

	return &token, nil
}

func (r *TokenInfoRepository) GetByRefresh(ctx context.Context, refresh string) (models.TokenInfo, error) {
	var token TokenInfo
	if err := r.db.QueryRowxContext(ctx, tokenGetByRefresh, refresh).StructScan(&token); err != nil {
		return nil, fmt.Errorf("error executing token_repository::get_by_refresh in postgres: %w", err)
	}

	return &token, nil
}

func (r *TokenInfoRepository) RemoveByCode(ctx context.Context, code string) error {
	if res := r.db.QueryRowContext(ctx, tokenRemoveByCode, code); res.Err() != nil {
		return fmt.Errorf("error executing token_repository::delete_by_code in postgres: %w", res.Err())
	}
	return nil
}

func (r *TokenInfoRepository) RemoveByAccess(ctx context.Context, access string) error {
	if res := r.db.QueryRowContext(ctx, tokenRemoveByAccess, access); res.Err() != nil {
		return fmt.Errorf("error executing token_repository::delete_by_access in postgres: %w", res.Err())
	}
	return nil
}

func (r *TokenInfoRepository) RemoveByRefresh(ctx context.Context, refresh string) error {
	if res := r.db.QueryRowContext(ctx, tokenRemoveByRefresh, refresh); res.Err() != nil {
		return fmt.Errorf("error executing token_repository::delete_by_refresh in postgres: %w", res.Err())
	}
	return nil
}

var _ models.TokenInfo = (*TokenInfo)(nil)
var _ repository.OAuth2TokenRepository = (*TokenInfoRepository)(nil)
