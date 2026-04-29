package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()
var isConnected bool

func init() {
	rand.Seed(time.Now().UnixNano())
}

func InitRedis() *redis.Client {
	Client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		log.Println("Redis连接失败: " + err.Error() + "，将不使用缓存功能")
		isConnected = false
	} else {
		log.Println("Redis连接成功(●'◡'●)")
		isConnected = true
	}

	return Client
}

func Delete(key string) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	return Client.Del(Ctx, key).Err()
}
func SetWithExpiration(key string, value interface{}, expiration time.Duration) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	return Client.Set(Ctx, key, value, expiration).Err()
}

func DeleteByPattern(pattern string) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	iter := Client.Scan(Ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(Ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return Client.Del(Ctx, keys...).Err()
	}
	return nil
}

func Exists(key string) (bool, error) {
	if !isConnected {
		return false, fmt.Errorf("redis not connected")
	}
	result, err := Client.Exists(Ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func Set(key string, value interface{}) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	return Client.Set(Ctx, key, value, 0).Err()
}

func Incr(key string) (int64, error) {
	if !isConnected {
		return 0, fmt.Errorf("redis not connected")
	}
	return Client.Incr(Ctx, key).Result()
}

func Decr(key string) (int64, error) {
	if !isConnected {
		return 0, fmt.Errorf("redis not connected")
	}
	return Client.Decr(Ctx, key).Result()
}

func GetJSON(key string, dest interface{}) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	data, err := Client.Get(Ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func SetJSON(key string, value interface{}, expiration time.Duration) error {
	return SetJSONWithJitter(key, value, expiration, 0.3)
}

func SetJSONWithJitter(key string, value interface{}, baseExpiration time.Duration, jitterRatio float64) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	expiration := baseExpiration
	if jitterRatio > 0 {
		jitter := int64(float64(baseExpiration.Nanoseconds()) * jitterRatio * rand.Float64())
		expiration = time.Duration(rand.Int63n(jitter) + baseExpiration.Nanoseconds() - jitter/2)
	}
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}
	return Client.Set(Ctx, key, jsonData, expiration).Err()
}

const (
	NullValueMarker = "__NULL_CACHE_MARKER__"
	DefaultTTL      = 5 * time.Minute
	BloomFilterKey  = "bloom:filter:%s"
)

func SetNullCache(key string, ttl time.Duration) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	return Client.Set(Ctx, key, NullValueMarker, ttl).Err()
}

func IsNullCache(key string) (bool, error) {
	if !isConnected {
		return false, fmt.Errorf("redis not connected")
	}
	val, err := Client.Get(Ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == NullValueMarker, nil
}

type SetNXResult struct {
	Success bool
	Err     error
}

func SetNX(key string, value interface{}, expiration time.Duration) SetNXResult {
	if !isConnected {
		return SetNXResult{false, fmt.Errorf("redis not connected")}
	}
	success, err := Client.SetNX(Ctx, key, value, expiration).Result()
	return SetNXResult{success, err}
}

type BloomFilter struct {
	name      string
	size      uint64
	hashFuncs int
}

func NewBloomFilter(name string, size uint64, hashFuncs int) *BloomFilter {
	return &BloomFilter{
		name:      name,
		size:      size,
		hashFuncs: hashFuncs,
	}
}

func (bf *BloomFilter) generateHashes(data string) []uint64 {
	hash1 := fnv64a(data)
	hash2 := fnv64a(data + "salt")
	hashes := make([]uint64, bf.hashFuncs)
	for i := 0; i < bf.hashFuncs; i++ {
		hashes[i] = (hash1 + uint64(i)*hash2) % bf.size
	}
	return hashes
}

func fnv64a(s string) uint64 {
	hash := uint64(2166136261)
	for _, c := range s {
		hash ^= uint64(c)
		hash *= 16777619
	}
	return hash
}

func (bf *BloomFilter) Add(key string) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	hashes := bf.generateHashes(key)
	pipe := Client.Pipeline()
	for _, h := range hashes {
		pipe.SetBit(Ctx, fmt.Sprintf(BloomFilterKey, bf.name), int64(h), 1)
	}
	_, err := pipe.Exec(Ctx)
	return err
}

func (bf *BloomFilter) Contains(key string) (bool, error) {
	if !isConnected {
		return false, fmt.Errorf("redis not connected")
	}
	hashes := bf.generateHashes(key)
	for _, h := range hashes {
		bit, err := Client.GetBit(Ctx, fmt.Sprintf(BloomFilterKey, bf.name), int64(h)).Result()
		if err != nil {
			return false, err
		}
		if bit == 0 {
			return false, nil
		}
	}
	return true, nil
}

func (bf *BloomFilter) Clear() error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	return Client.Del(Ctx, fmt.Sprintf(BloomFilterKey, bf.name)).Err()
}

var (
	UserInfoFilter *BloomFilter
	VideoFilter    *BloomFilter
	CommentFilter  *BloomFilter
)

func InitBloomFilters() {
	UserInfoFilter = NewBloomFilter("user_info", 100000, 3)
	VideoFilter = NewBloomFilter("video", 100000, 3)
	CommentFilter = NewBloomFilter("comment", 100000, 3)
}

func AddToBloomFilter(filterType string, key string) error {
	var filter *BloomFilter
	switch filterType {
	case "user":
		filter = UserInfoFilter
	case "video":
		filter = VideoFilter
	case "comment":
		filter = CommentFilter
	default:
		return fmt.Errorf("unknown filter type: %s", filterType)
	}
	if filter == nil {
		return nil
	}
	return filter.Add(key)
}

func ExistsInBloomFilter(filterType string, key string) (bool, error) {
	var filter *BloomFilter
	switch filterType {
	case "user":
		filter = UserInfoFilter
	case "video":
		filter = VideoFilter
	case "comment":
		filter = CommentFilter
	default:
		return false, fmt.Errorf("unknown filter type: %s", filterType)
	}
	if filter == nil {
		return false, nil
	}
	return filter.Contains(key)
}
