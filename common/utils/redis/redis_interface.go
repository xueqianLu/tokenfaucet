package redis
type IRedis interface {
	Set(key, value string) error
	SetBytes(key string, value []byte) error
	SetKvAndExp(key, value string, expire int) error
	Get(key string) string
	GetBytes(key string) []byte
	IsKeyExists(key string) bool
	Del(key string) bool
	Setnx(key string, value string) error
	Lpush(key string, value ...int) error
	LpushCount(key string, number int) error
	LpushByte(key string, value []byte) error
	Close()
	LPop(key string) (string, error)
	LLen(key string) (int64, error)
	HINCRBY(key, field string)
	HGet(key, field string) (interface{}, error)
	HGetAll(key string) ([][]byte, error)
	HSet(key string, field string, value interface{}) (interface{}, error)
}
