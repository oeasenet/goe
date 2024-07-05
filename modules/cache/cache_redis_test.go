package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRedisCache(t *testing.T) {
	// 模拟 Redis 连接配置
	redisHost := "192.168.30.1"
	redisPort := 6379
	redisUsername := ""
	redisPassword := "windowsbcldbb"
	redisDB := 0

	// 调用函数
	rc := NewRedisCache(redisHost, redisPort, redisUsername, redisPassword, redisDB)

	// 检查返回的 RedisCache 是否不为 nil
	assert.NotNil(t, rc, "RedisCache should not be nil")

	// 检查 store 属性是否不为 nil
	assert.NotNil(t, rc.store, "store should not be nil")
}

func TestSetAndGet(t *testing.T) {
	redisHost := "192.168.30.1"
	redisPort := 6379
	redisUsername := ""
	redisPassword := "windowsbcldbb"
	redisDB := 0

	rc := NewRedisCache(redisHost, redisPort, redisUsername, redisPassword, redisDB)

	var value = []byte("doe")
	err := rc.Set("testKey", value, 0)

	assert.NoError(t, err, "Should not return an error")

	data := rc.Get("testKey")
	assert.Equal(t, data, value, "Stored value should match the set value")
}

func TestSetExpiration(t *testing.T) {
	redisHost := "192.168.30.1"
	redisPort := 6379
	redisUsername := ""
	redisPassword := "windowsbcldbb"
	redisDB := 0

	rc := NewRedisCache(redisHost, redisPort, redisUsername, redisPassword, redisDB)

	var value = []byte("doe")
	var exp = 1 * time.Second

	err := rc.Set("testKey", value, exp)
	assert.NoError(t, err, "Should not return an error")

	time.Sleep(1100 * time.Millisecond)

	data := rc.Get("testKey")
	assert.Nil(t, data, "Stored value should match the set value")

}

func TestSetBingAndGet(t *testing.T) {
	redisHost := "192.168.30.1"
	redisPort := 6379
	redisUsername := ""
	redisPassword := "windowsbcldbb"
	redisDB := 0

	rc := NewRedisCache(redisHost, redisPort, redisUsername, redisPassword, redisDB)

	type person struct {
		Age  int    `json:"age"`
		Name string `json:"name"`
	}

	var test = person{
		Age:  23,
		Name: "tester",
	}

	err := rc.SetBind("testKey", &test, 0)
	assert.NoError(t, err, "Should not return an error")

	var data person
	err = rc.GetBind("testKey", &data)
	assert.NoError(t, err, "Should not return an error")

	assert.Equal(t, data, test, "Stored value should match the set value")
}

func TestDelete(t *testing.T) {
	redisHost := "192.168.30.1"
	redisPort := 6379
	redisUsername := ""
	redisPassword := "windowsbcldbb"
	redisDB := 0

	rc := NewRedisCache(redisHost, redisPort, redisUsername, redisPassword, redisDB)

	var value = []byte("doe")

	err := rc.Set("testKey", value, 0)
	assert.NoError(t, err, "Should not return an error")

	err = rc.Delete("testKey")
	assert.NoError(t, err, "Should not return an error")

	data := rc.Get("testKey")
	assert.Nil(t, data, "Stored value should be nil")
}
