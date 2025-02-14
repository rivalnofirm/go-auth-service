package history

import (
	"github.com/jmoiron/sqlx"
	"log"

	"go-auth-service/src/infra/constants/common"
	"go-auth-service/src/infra/models"
	"go-auth-service/src/infra/persistence/postgres"
)

type HistoryRepository interface {
	Create(userId int64, ipAddress, userAgent string) error
	UpdateLogoutByUserIdAndUserAgent(userId int64, logoutReason, userAgent string) error
	GetByUserId(useId int64) ([]*models.History, error)
	UpdateLogoutByUserId(userId int64, logoutReason string) error
}

const (
	Create                           = `INSERT INTO user_login_history (user_id, login_time, ip_address, user_agent) VALUES ($1, now(), $2, $3)`
	UpdateLogoutByUserIdAndUserAgent = `UPDATE user_login_history SET logout_time = NOW(), logout_reason = $1 WHERE user_id = $2 AND user_agent = $3 AND logout_time IS NULL`
	GetByUserId                      = `SELECT * FROM user_login_history WHERE user_id = $1 AND logout_time IS NULL ORDER BY login_time`
	UpdateLogoutByUserId             = "UPDATE user_login_history SET logout_time = NOW(), logout_reason = $1 WHERE user_id = $2 AND logout_time IS NULL"
)

type PreparedStatement struct {
	create                           *sqlx.Stmt
	updateLogoutByUserIdAndUserAgent *sqlx.Stmt
	getByUserId                      *sqlx.Stmt
	updateLogoutByUserId             *sqlx.Stmt
}

type historyRepo struct {
	Connection *postgres.Connection
	statement  PreparedStatement
}

func NewHistoryRepository(db *postgres.Connection) HistoryRepository {
	repo := &historyRepo{
		Connection: db,
	}
	InitPreparedStatement(repo)
	return repo
}

func (p *historyRepo) Preparex(query string, isMaster bool) *sqlx.Stmt {
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

func InitPreparedStatement(m *historyRepo) {
	m.statement = PreparedStatement{
		create:                           m.Preparex(Create, common.IsMasterDb),
		updateLogoutByUserIdAndUserAgent: m.Preparex(UpdateLogoutByUserIdAndUserAgent, common.IsMasterDb),
		getByUserId:                      m.Preparex(GetByUserId, common.NotIsMasterDb),
		updateLogoutByUserId:             m.Preparex(UpdateLogoutByUserId, common.IsMasterDb),
	}
}

func (p *historyRepo) Create(userId int64, ipAddress, userAgent string) error {
	_, err := p.statement.create.Exec(userId, ipAddress, userAgent)
	if err != nil {
		return err
	}

	return nil
}

func (p *historyRepo) UpdateLogoutByUserIdAndUserAgent(userId int64, logoutReason, userAgent string) error {
	_, err := p.statement.updateLogoutByUserIdAndUserAgent.Exec(logoutReason, userId, userAgent)
	if err != nil {
		return err
	}

	return nil
}

func (p *historyRepo) GetByUserId(userId int64) ([]*models.History, error) {
	var history []*models.History

	err := p.statement.getByUserId.Select(&history, userId)
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return []*models.History{}, nil
	}

	return history, nil
}

func (p *historyRepo) UpdateLogoutByUserId(userId int64, logoutReason string) error {
	_, err := p.statement.updateLogoutByUserId.Exec(logoutReason, userId)
	if err != nil {
		return err
	}

	return nil
}
