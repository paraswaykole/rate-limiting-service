package storage

type StorageManager struct {
	RedisStorage *RedisStorage
}

var storageManager *StorageManager

func GetManager() *StorageManager {
	if storageManager == nil {
		storageManager = &StorageManager{
			RedisStorage: InitRedisStorage(),
		}
	}
	return storageManager
}
