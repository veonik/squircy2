package squircy

import (
	"github.com/fzzy/radix/redis"
)

func newRedisClient(config *Configuration) (c *redis.Client) {
	c, err := redis.Dial("tcp", config.RedisHost)
	if err != nil {
		panic("Error connecting to redis")
	}
	
	r := c.Cmd("select", config.RedisDatabase)
	if r.Err != nil {
		panic("Error selecting redis database")
	}
	
	return
}