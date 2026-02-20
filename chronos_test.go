package Chronos

import (
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestChronos(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 有密码就填
		DB:       0,
	})
	for i := 0; i < 5; i++ {
		go func(i int) {
			chronos := New(rdb)
			chronos.AddDistributedTask("@every 2s", func() {
				print(i)
				fmt.Sprintf("我是%d", i)
			}, Option{
				TaskName: "测试",
				ExecTime: time.Second * 2,
			})
			chronos.Start()
		}(i)
	}
	time.Sleep(time.Second * 100)
}
