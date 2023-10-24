// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

// Queue is Queue instance.
type Queue struct {
	opt redis.Options
	cli *Client
	svr *asynq.Server
}

// Client queue client.
type Client struct {
	cli *asynq.Client
	hds *sync.Map
}

// New returns a Queue instance.
func New(opt redis.Options) *Queue {
	q := &Queue{opt: opt}

	o := &asynq.RedisClientOpt{
		Network:      opt.Network,
		Addr:         opt.Addr,
		Username:     opt.Username,
		Password:     opt.Password,
		DB:           opt.DB,
		DialTimeout:  opt.DialTimeout,
		ReadTimeout:  opt.ReadTimeout,
		WriteTimeout: opt.WriteTimeout,
		PoolSize:     opt.PoolSize,
		TLSConfig:    opt.TLSConfig,
	}

	q.cli = &Client{
		cli: asynq.NewClient(o),
		hds: new(sync.Map),
	}

	conf := asynq.Config{
		Concurrency: 4,
		Logger:      log.StandardLogger(),
	}
	conf.LogLevel.Set(log.StandardLogger().Level.String())
	if conf.LogLevel < asynq.InfoLevel {
		conf.LogLevel = asynq.InfoLevel // at least level info
	}

	q.svr = asynq.NewServer(o, conf)

	return q
}

// Run tasks with handlers.
// Special attention, please run after registration is completed.
func (q *Queue) Run() {
	go func() {
		f := func(ctx context.Context, task *asynq.Task) error {
			v, _ := q.cli.hds.Load(task.Type())
			h, _ := v.(asynq.Handler)
			if h == nil {
				log.WithField("task_type", task.Type()).Error("no handler")
				return fmt.Errorf("no handler for task type: %s", task.Type())
			}
			if log.IsLevelEnabled(log.DebugLevel) {
				log.WithFields(map[string]any{"task_type": task.Type(), "task_payload": task.Payload()}).Debugf("receive task from queue")
			}
			return h.ProcessTask(ctx, task)
		}

		if err := q.svr.Run(asynq.HandlerFunc(f)); err != nil {
			log.Fatal("run handler error:", err)
		}
	}()
}

// Client returns the queue client.
func (q *Queue) Client() *Client {
	return q.cli
}

// Enqueue enqueues the given task and payload to a queue.
func (q *Client) Enqueue(taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	t := asynq.NewTask(taskType, payload, opts...)
	return q.cli.Enqueue(t)
}

// RegisterHandler register handler for the task type.
func (q *Client) RegisterHandler(taskType string, handler asynq.Handler) error {
	if _, ok := q.hds.Load(taskType); ok {
		return errors.New("task type already registered")
	}
	q.hds.Store(taskType, handler)
	return nil
}
