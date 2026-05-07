package cron

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
)

// TaskBuilder 任务构建器
type TaskBuilder struct {
	name   string
	spec   string
	hour   uint
	minute uint
	task   func()
}

var globalScheduler *gocron.Scheduler

// InitScheduler 初始化 Cron 调度器
func InitScheduler() {
	globalScheduler = gocron.NewScheduler(time.Local)
	log.Println("初始化 Cron 调度器成功")
}

// StartScheduler 启动 Cron 调度器
func StartScheduler() {
	if globalScheduler == nil {
		log.Println("Cron 调度器没有初始化,请先初始化调度器")
	}
	globalScheduler.StartAsync()
	log.Println("已启动 Cron 调度器")
}

// StopScheduler 停止 Cron 调度器
func StopScheduler() {
	if globalScheduler == nil {
		return
	}
	globalScheduler.Stop()
	log.Println("已停止计划任务")
}

// Register 通过任务名称返回一个任务构造器
func Register(name string) *TaskBuilder {
	return &TaskBuilder{name: name}
}

// Daily 设置每日固定时间执行
func (b *TaskBuilder) Daily(hour, minute uint) *TaskBuilder {
	b.hour = hour
	b.minute = minute
	return b
}

// Cron 设置 Cron 表达式 ( 优先级高于 Daily )
func (b *TaskBuilder) Cron(spec string) *TaskBuilder {
	b.spec = spec
	return b
}

func (b *TaskBuilder) Do(task func()) {
	if globalScheduler == nil {
		log.Fatalf("Cron 调度器未初始化,请先调用初始化调度器")
	}
	if task == nil {
		log.Fatalf("任务 [%s] 没有提供执行函数", b.name)
	}
	b.task = task

	var err error
	if b.spec != "" {
		_, err = globalScheduler.Cron(b.spec).Do(b.task)
		if err == nil {
			log.Printf("注册计划任务成功: %s -> %s", b.name, b.spec)
		}
	} else {
		timeStr := fmt.Sprintf("%02d:%02d", b.hour, b.minute)
		_, err = globalScheduler.Every(1).Day().At(timeStr).Do(b.task)
		if err == nil {
			log.Printf("注册每日任务成功: %s -> %02d:%02d", b.name, b.hour, b.minute)
		}
	}
	if err != nil {
		log.Fatalf("注册任务 [%s] 失败: %w", b.name, err)
	}
}
