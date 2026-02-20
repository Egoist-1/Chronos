package Chronos

import (
	"fmt"
	"time"
)

func main() {
	chronos := New(nil)
	chronos.AddDistributedTask("@every 2s", func() {
		fmt.Println("任务")
	}, Option{
		TaskName: "测试",
		ExecTime: time.Second * 2,
	})
	chronos.Start()
	stop := chronos.Stop()
	<-stop.Done()
}
