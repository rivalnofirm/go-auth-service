package config

import (
	"os"
	"strconv"
)

type AppConf struct {
	Environment string
	Name        string
}

type HttpConf struct {
	Port       string
	XRequestID string
	Timeout    int
}

type LogConf struct {
	Name string
}

type SqlDbInstanceConf struct {
	Host                   string
	Username               string
	Password               string
	Name                   string
	Port                   string
	SSLMode                string
	Schema                 string
	MaxOpenConn            int
	MaxIdleConn            int
	MaxIdleTimeConnSeconds int64
	MaxLifeTimeConnSeconds int64
}

type SqlDbConf struct {
	Master SqlDbInstanceConf
	Slave  SqlDbInstanceConf
}

type RedisConf struct {
	Host string
	Port string
}

type Config struct {
	App   AppConf
	Http  HttpConf
	Log   LogConf
	SqlDb SqlDbConf
	Redis RedisConf
}

func Make() Config {
	app := AppConf{
		Environment: os.Getenv("APP_ENV"),
		Name:        os.Getenv("APP_NAME"),
	}

	master := SqlDbInstanceConf{
		Host:     os.Getenv("DB_MASTER_HOST"),
		Username: os.Getenv("DB_MASTER_USERNAME"),
		Password: os.Getenv("DB_MASTER_PASSWORD"),
		Name:     os.Getenv("DB_MASTER_NAME"),
		Port:     os.Getenv("DB_MASTER_PORT"),
		SSLMode:  os.Getenv("DB_MASTER_SSL_MODE"),
		Schema:   os.Getenv("DB_MASTER_SCHEMA"),
	}

	slave := SqlDbInstanceConf{
		Host:     os.Getenv("DB_SLAVE_HOST"),
		Username: os.Getenv("DB_SLAVE_USERNAME"),
		Password: os.Getenv("DB_SLAVE_PASSWORD"),
		Name:     os.Getenv("DB_SLAVE_NAME"),
		Port:     os.Getenv("DB_SLAVE_PORT"),
		SSLMode:  os.Getenv("DB_SLAVE_SSL_MODE"),
		Schema:   os.Getenv("DB_SLAVE_SCHEMA"),
	}

	redis := RedisConf{
		Host: os.Getenv("REDIS_HOST"),
		Port: os.Getenv("REDIS_PORT"),
	}

	http := HttpConf{
		Port:       os.Getenv("HTTP_PORT"),
		XRequestID: os.Getenv("HTTP_REQUEST_ID"),
	}

	httpTimeout, err := strconv.Atoi(os.Getenv("HTTP_TIMEOUT"))
	if err == nil {
		http.Timeout = httpTimeout
	}

	config := Config{
		App:  app,
		Http: http,
		Log: LogConf{
			Name: os.Getenv("LOG_NAME"),
		},
		SqlDb: SqlDbConf{
			Master: master,
			Slave:  slave,
		},
		Redis: redis,
	}

	return config
}
