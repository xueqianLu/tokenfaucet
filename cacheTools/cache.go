package cacheTools

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	"github.com/hpb-project/tokenfaucet/common"
	red "github.com/hpb-project/tokenfaucet/common/utils/redis"
	"strconv"
	"time"
)

var SR *red.StoreRedis
var pool *redis.Pool

func init() {
	conf := common.GetCacheConfig()
	NewPool(conf.Conn, conf.DBNum, conf.Passworkd)
	SR = &red.StoreRedis{}
	SR.SetPool(pool)
}

func NewPool(conn, dbNum, password string) {
	pool = &redis.Pool{
		MaxIdle:     50, //最大空闲连接数
		MaxActive:   0,  //若为0，则活跃数没有限制
		Wait:        true,
		IdleTimeout: 30 * time.Second, //最大空闲连接时间
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conn)
			if err != nil {
				logs.Error(err)
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			c.Do("SELECT", dbNum)
			logs.Info("redis connect successful")
			return c, nil
		},
	}

}

func CheckIpAddress(ip string) (int, string, error) {
	var count string //计数
	var time int     //key 的过期时间
	c := "1"
	time = common.GetLimitTime()
	ct := common.GetLimitCount()

	if boo := SR.IsKeyExists(ip); boo {
		count = SR.Get(ip)
		time = SR.GetExp(ip)
		if count == ct {
			return 0, "", fmt.Errorf("each IP has only %s chance every 24 hours", ct)
		}
		num, _ := strconv.Atoi(count)
		c = strconv.Itoa(num + 1)
	}
	return time, c, nil
}
