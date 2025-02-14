package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"go-auth-service/src/infra/config"
)

type Connection struct {
	master dbConnection
	slave  dbConnection
}

type dbConnection struct {
	primary *sqlx.DB
	portal  map[string]*sqlx.DB
}

func NewConnection(masterConf, slaveConf config.SqlDbInstanceConf, logger *logrus.Logger) (*Connection, error) {
	conn := &Connection{
		master: dbConnection{portal: make(map[string]*sqlx.DB)},
		slave:  dbConnection{portal: make(map[string]*sqlx.DB)},
	}

	// Connect to Master
	masterDSN := buildDSN(masterConf)
	masterConn, err := sqlx.Connect("postgres", masterDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %v", err)
	}

	// Set maximum open connections
	masterConn.SetMaxOpenConns(masterConf.MaxOpenConn)
	masterConn.SetMaxIdleConns(masterConf.MaxIdleConn)
	masterConn.SetConnMaxIdleTime(time.Duration(masterConf.MaxIdleTimeConnSeconds) * time.Second)
	masterConn.SetConnMaxLifetime(time.Duration(masterConf.MaxLifeTimeConnSeconds) * time.Second)

	// Ping master connection
	err = masterConn.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping master database: %v", err)
	}

	conn.master.primary = masterConn

	// Connect to Slave
	slaveDSN := buildDSN(slaveConf)
	slaveConn, err := sqlx.Connect("postgres", slaveDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to slave database: %v", err)
	}

	// Set maximum open connections
	slaveConn.SetMaxOpenConns(slaveConf.MaxOpenConn)
	slaveConn.SetMaxIdleConns(slaveConf.MaxIdleConn)
	slaveConn.SetConnMaxIdleTime(time.Duration(slaveConf.MaxIdleTimeConnSeconds) * time.Second)
	slaveConn.SetConnMaxLifetime(time.Duration(slaveConf.MaxLifeTimeConnSeconds) * time.Second)

	// Ping slave connection
	err = slaveConn.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping slave database: %v", err)
	}

	conn.slave.primary = slaveConn

	logger.Println("PostgreSQL master-slave database connection established")

	return conn, nil
}

func buildDSN(conf config.SqlDbInstanceConf) string {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		conf.Host,
		conf.Username,
		conf.Password,
		conf.Name,
		conf.Port,
		conf.SSLMode,
	)
	if conf.Schema != "" {
		dsn += fmt.Sprintf(" search_path=%s", conf.Schema)
	}
	return dsn
}

func (c *Connection) GetPrimaryMaster() *sqlx.DB {
	return c.master.primary
}

func (c *Connection) GetPrimarySlave() *sqlx.DB {
	return c.slave.primary
}

func (c *Connection) GetPortalMaster(portal string) *sqlx.DB {
	portal = strings.ToLower(portal)
	if _, ok := c.master.portal[portal]; ok {
		return c.master.portal[portal]
	}
	return c.master.primary
}

func (c *Connection) GetPortalSlave(portal string) *sqlx.DB {
	portal = strings.ToLower(portal)
	if _, ok := c.slave.portal[portal]; ok {
		return c.slave.portal[portal]
	}
	return c.slave.primary
}
