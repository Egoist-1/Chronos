package Chronos

import (
	"context"
	"log"
	"sync"
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

func New(client redis.Cmdable) *Chronos {
	return &Chronos{
		cron.New(),
		rlock.NewClient(client),
	}
}

type Chronos struct {
	*cron.Cron
	r *rlock.Client
}

func (c *Chronos) AddDistributedTask(spec string, cmd func(), option Option) (cron.EntryID, error) {
	task := &Task{
		Option: option,
		r:      c.r,
	}
	job := task.newTask(cmd)
	return c.AddFunc(spec, job)
}

type Option struct {
	TaskName string        //也是redis锁的key
	ExecTime time.Duration //预期执行时间
}

type Task struct {
	Option
	r         *rlock.Client
	lock      *rlock.Lock
	localLock sync.Mutex
}

func (t *Task) newTask(job func()) func() {
	return func() {
		//没有锁,尝试去拿锁
		if t.lock == nil {
			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
			defer cancelFunc()
			//第一个ctx 是拿锁的过期时间
			lock, err := t.r.Lock(ctx, t.TaskName, t.ExecTime, &rlock.FixIntervalRetry{
				Interval: time.Second,
				Max:      2,
			}, time.Second)
			if err != nil {
				return
			}
			t.lock = lock
			//拿到锁之后,开始续约
			go func() {
				t.localLock.Lock()
				defer t.localLock.Unlock()
				er := lock.AutoRefresh(t.ExecTime/2, time.Second)
				if er != nil {
					log.Print("续约失败")
				}
				t.lock = nil
			}()
		}
		job()
	}
}
