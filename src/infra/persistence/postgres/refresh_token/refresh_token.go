package refresh_token

import (
	"errors"
	"github.com/jmoiron/sqlx"
	"go-auth-service/src/infra/constants/common"
	errorMessage "go-auth-service/src/infra/constants/error_message"
	"go-auth-service/src/infra/models"
	"log"
	"time"

	"go-auth-service/src/infra/persistence/postgres"
)

type RefreshTokenRepository interface {
	Create(userId int64, refreshTokenHash, userAgent string) error
	GetTokenActive(userId int64, userAgent string) (*models.UserRefreshToken, error)
	UpdateStatus(userId int64, userAgent string) error
	UpdateStatusByUserId(userId int64) error
}

const (
	Create               = `INSERT INTO user_refresh_token (user_id, refresh_token_hash, expires_at, user_agent) VALUES ($1, $2, $3, $4)`
	GetTokenActive       = `SELECT * FROM user_refresh_token WHERE user_id = $1 AND user_agent = $2 AND is_active = TRUE`
	UpdateStatus         = `UPDATE user_refresh_token SET is_active = FALSE WHERE user_id = $1 AND user_agent = $2`
	UpdateStatusByUserId = `UPDATE user_refresh_token SET is_active = FALSE WHERE user_id = $1`
)

type PreparedStatement struct {
	create               *sqlx.Stmt
	getTokenActive       *sqlx.Stmt
	updateStatus         *sqlx.Stmt
	updateStatusByUserId *sqlx.Stmt
}

type refreshTokenRepo struct {
	Connection *postgres.Connection
	statement  PreparedStatement
}

func NewRefreshTokenRepository(db *postgres.Connection) RefreshTokenRepository {
	repo := &refreshTokenRepo{
		Connection: db,
	}
	InitPreparedStatement(repo)
	return repo
}

func (p *refreshTokenRepo) Preparex(query string, isMaster bool) *sqlx.Stmt {
	if !isMaster {
		// for slave
		statement, err := p.Connection.GetPrimarySlave().Preparex(query)
		if err != nil {
			log.Fatalf("Failed to preparex query: %s. Error: %s", query, err.Error())
		}

		return statement
	}

	statement, err := p.Connection.GetPrimaryMaster().Preparex(query)
	if err != nil {
		log.Fatalf("Failed to preparex query: %s. Error: %s", query, err.Error())
	}

	return statement
}

func InitPreparedStatement(m *refreshTokenRepo) {
	m.statement = PreparedStatement{
		create:               m.Preparex(Create, common.IsMasterDb),
		getTokenActive:       m.Preparex(GetTokenActive, common.NotIsMasterDb),
		updateStatus:         m.Preparex(UpdateStatus, common.IsMasterDb),
		updateStatusByUserId: m.Preparex(UpdateStatusByUserId, common.IsMasterDb),
	}
}

func (p *refreshTokenRepo) Create(userId int64, refreshTokenHash, userAgent string) error {
	expiresAt := time.Now().Add(common.RefreshTokenExp)
	_, err := p.statement.create.Exec(userId, refreshTokenHash, expiresAt, userAgent)
	if err != nil {
		return err
	}

	return nil
}

func (p *refreshTokenRepo) GetTokenActive(userId int64, userAgent string) (*models.UserRefreshToken, error) {
	var refreshToken []*models.UserRefreshToken

	err := p.statement.getTokenActive.Select(&refreshToken, userId, userAgent)
	if err != nil {
		return nil, err
	}

	if len(refreshToken) < 1 {
		return nil, errors.New(errorMessage.UserRefreshTokenNotFound)
	}

	return refreshToken[0], nil
}

func (p *refreshTokenRepo) UpdateStatus(userId int64, userAgent string) error {
	_, err := p.statement.updateStatus.Exec(userId, userAgent)
	if err != nil {
		return err
	}

	return nil
}

func (p *refreshTokenRepo) UpdateStatusByUserId(userId int64) error {
	_, err := p.statement.updateStatusByUserId.Exec(userId)
	if err != nil {
		return err
	}

	return nil
}
