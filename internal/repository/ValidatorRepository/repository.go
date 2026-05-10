package ValidatorRepository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Interface interface {
	AcquireCooldown(to string, ttl time.Duration) (bool, error)
	StoreCode(to, code string, ttl time.Duration) error
	GetCode(to string) (string, error)
	DeleteCooldown(to string)
	DeleteCode(to string)
}

type ValidatorRepository struct {
	client *redis.Client
}

func NewValidatorRepository(client *redis.Client) *ValidatorRepository {
	return &ValidatorRepository{client: client}
}

func (r *ValidatorRepository) AcquireCooldown(to string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(context.Background(), "verify:cooldown:"+to, 1, ttl).Result()
}

func (r *ValidatorRepository) StoreCode(to, code string, ttl time.Duration) error {
	return r.client.Set(context.Background(), "verify:code:"+to, code, ttl).Err()
}

func (r *ValidatorRepository) GetCode(to string) (string, error) {
	return r.client.Get(context.Background(), "verify:code:"+to).Result()
}

func (r *ValidatorRepository) DeleteCooldown(to string) {
	r.client.Del(context.Background(), "verify:cooldown:"+to)
}

func (r *ValidatorRepository) DeleteCode(to string) {
	r.client.Del(context.Background(), "verify:code:"+to)
}
