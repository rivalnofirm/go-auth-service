package service

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type ServiceRedis struct {
	Rdb *redis.Client
}

type ServRedisInterface interface {
	SetData(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetData(ctx context.Context, key string) (string, error)
	DeleteData(ctx context.Context, key string) error
	IsAllowed(ctx context.Context, key string, limit int, duration time.Duration) (bool, error)
}

func NewServRedis(rdb *redis.Client) *ServiceRedis {
	return &ServiceRedis{
		Rdb: rdb,
	}
}

// SetData menyimpan data di Redis dengan TTL (Time to Live)
func (p *ServiceRedis) SetData(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	err := p.Rdb.Set(ctx, key, value, ttl).Err()
	if err != nil {
		log.Println("redis set failed:", err)
		return err
	}
	return nil
}

// GetData mengambil data dari Redis berdasarkan key
func (p *ServiceRedis) GetData(ctx context.Context, key string) (string, error) {
	dataRedis, err := p.Rdb.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Failed to get data from redis for key %s: %v", key, err)
		return "", err
	}
	return dataRedis, nil
}

func (p *ServiceRedis) DeleteData(ctx context.Context, key string) error {
	err := p.Rdb.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Failed to delete data from redis for key %s: %v", key, err)
		return err
	}
	log.Printf("Successfully deleted data from redis for key %s", key)
	return nil
}

func (p *ServiceRedis) IsAllowed(ctx context.Context, key string, limit int, duration time.Duration) (bool, error) {
	count, err := p.Rdb.Incr(ctx, key).Result()
	if err != nil {
		log.Println("Error incrementing rate limit:", err)
		return false, err
	}

	if count == 1 {
		p.Rdb.Expire(ctx, key, duration)
	}

	if count > int64(limit) {
		return false, nil
	}

	return true, nil
}
