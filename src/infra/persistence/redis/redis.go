package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"go-auth-service/src/infra/config"
)

func NewRedisClient(conf config.RedisConf, logger *logrus.Logger) (*redis.Client, error) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Host + ":" + conf.Port,
		Password: "",
		DB:       0,
	})
	_, err := client.Ping(ctx).Result()

	if err != nil {
		logger.Printf("cant connect Redis :  %s", err)
	}

	return client, nil
}
