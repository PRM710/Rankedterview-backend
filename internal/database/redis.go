package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Z represents a sorted set member with score
type Z = redis.Z

// RedisClient wraps the Redis client
type RedisClient struct {
	Client *redis.Client
}

// NewRedis creates a new Redis connection
func NewRedis(addr, password string, db int) *RedisClient {
	// Use direct configuration with address, password, and db
	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	client := redis.NewClient(opts)

	return &RedisClient{
		Client: client,
	}
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// Ping checks if Redis is reachable
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Set sets a key-value pair with optional expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Del deletes one or more keys
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	return result > 0, err
}

// HSet sets a field in a hash
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.HSet(ctx, key, values...).Err()
}

// HGet gets a field from a hash
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields from a hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// ZAdd adds members to a sorted set
func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.Client.ZAdd(ctx, key, members...).Err()
}

// ZRange retrieves a range from a sorted set
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange retrieves a range from a sorted set in reverse order
func (r *RedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRank gets the rank of a member in a sorted set
func (r *RedisClient) ZRank(ctx context.Context, key, member string) (int64, error) {
	return r.Client.ZRank(ctx, key, member).Result()
}

// ZRevRank gets the reverse rank of a member
func (r *RedisClient) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	return r.Client.ZRevRank(ctx, key, member).Result()
}

// ZScore gets the score of a member
func (r *RedisClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	return r.Client.ZScore(ctx, key, member).Result()
}

// Publish publishes a message to a channel
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}

// SAdd adds members to a set
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SAdd(ctx, key, members...).Err()
}

// SMembers gets all members of a set
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

// SRem removes members from a set
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SRem(ctx, key, members...).Err()
}

// Expire sets an expiration on a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}
