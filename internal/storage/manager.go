package storage

import (
	"context"
	"errors"
	"fmt"
	"rate-limiting-service/internal/utils"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type StorageManager struct {
	redisStorage *RedisStorage
}

var storageManager *StorageManager

func GetManager() *StorageManager {
	if storageManager == nil {
		storageManager = &StorageManager{
			redisStorage: InitRedisStorage(),
		}
	}
	return storageManager
}

func (sm *StorageManager) GetLimiterData(key string, out any) error {
	data, err := sm.redisStorage.client.HGetAll(context.Background(), key).Result()
	if err != nil && err.Error() == "redis: nil" {
		return errors.New(ErrDataNotFound)
	}
	if len(data) == 0 {
		return errors.New(ErrDataNotFound)
	}
	if err != nil {
		return err
	}
	return utils.MapToStruct(data, out)
}

func (sm *StorageManager) GetLimiterField(key string, field string) (string, error) {
	data, err := sm.redisStorage.client.HGet(context.Background(), key, field).Result()
	if err != nil && err.Error() == "redis: nil" {
		return "", errors.New(ErrDataNotFound)
	}
	if err != nil {
		return "", err
	}
	return data, nil
}

func (sm *StorageManager) SetLimiterData(key string, data any, ttlInSeconds int) {
	ttl := time.Second * time.Duration(ttlInSeconds)
	values := utils.StructToMap(data)
	err := sm.redisStorage.client.HSet(context.Background(), key, values).Err()
	if err != nil {
		panic(err)
	}
	sm.redisStorage.client.Expire(context.Background(), key, ttl)
}

func (sm *StorageManager) GetConfigureData(key string, out any) error {
	storageKey := fmt.Sprintf("configure:%s", key)
	data, err := sm.redisStorage.client.HGetAll(context.Background(), storageKey).Result()
	if err != nil && err.Error() == "redis: nil" {
		return errors.New(ErrDataNotFound)
	}
	if len(data) == 0 {
		return errors.New(ErrDataNotFound)
	}
	if err != nil {
		return err
	}
	return utils.MapToStruct(data, out)
}

func (sm *StorageManager) GetConfigureType(key string) (int, error) {
	storageKey := fmt.Sprintf("configure:%s", key)
	data, err := sm.redisStorage.client.HGet(context.Background(), storageKey, CONFIGURATION_LIMITER_TYPE_KEY).Result()
	if err != nil && err.Error() == "redis: nil" {
		return 0, errors.New(ErrDataNotFound)
	}
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(data)
}

func (sm *StorageManager) SetConfigureData(key string, limiterType int, data any) {
	storageKey := fmt.Sprintf("configure:%s", key)
	values := utils.StructToMap(data)
	values[CONFIGURATION_LIMITER_TYPE_KEY] = limiterType
	err := sm.redisStorage.client.HSet(context.Background(), storageKey, values).Err()
	if err != nil {
		panic(err)
	}
}

func (sm *StorageManager) PublishUpdates(channel string, data any) {
	sm.redisStorage.client.Publish(context.Background(), channel, data)
}

func (sm *StorageManager) SubscribeUpdates(channel string) *redis.PubSub {
	return sm.redisStorage.client.Subscribe(context.Background(), channel)
}
