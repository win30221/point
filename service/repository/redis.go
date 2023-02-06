package repository

import (
	"context"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/win30221/core/storage/redis"
	"github.com/win30221/point/domain"
)

const (
	BALANCE   = "BALANCE:"
	User_LOGS = "User_LOGS:"
)

type RedisRepo struct {
	pool *redigo.Pool

	balanceTTL   int
	pointLogsTTL int
}

func NewRedisRepo(pool *redigo.Pool, balanceTTL, pointLogsTTL int) *RedisRepo {
	return &RedisRepo{pool, balanceTTL, pointLogsTTL}
}

func (r *RedisRepo) GetBalance(ctx context.Context, userID string) (balance float64, err error) {
	err = redis.GET(r.pool, ctx, BALANCE+userID, &balance)
	return
}

func (r *RedisRepo) SetBalance(ctx context.Context, userID string, balance float64) (err error) {
	err = redis.SETEX(r.pool, ctx, BALANCE+userID, r.balanceTTL, balance)
	return
}

func (r *RedisRepo) GetUserPointLogs(ctx context.Context, userID string) (result []domain.PointLog, err error) {
	err = redis.GET(r.pool, ctx, User_LOGS+userID, &result)
	return
}

func (r *RedisRepo) SetUserPointLogs(ctx context.Context, userID string, result []domain.PointLog) (err error) {
	err = redis.SETEX(r.pool, ctx, User_LOGS+userID, r.pointLogsTTL, result)
	return
}
