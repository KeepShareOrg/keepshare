// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	mysqlDB  *gorm.DB
	redisCli *redis.Client
)

// MySQL returns the mysql client instance.
func MySQL() *gorm.DB {
	if mysqlDB == nil {
		log.Fatal("mysql has not been initialized")
	}
	return mysqlDB
}

// Redis returns the redis client instance.
func Redis() *redis.Client {
	if redisCli == nil {
		log.Fatal("redis has not been initialized")
	}
	return redisCli
}

func initMysql() error {
	dsn := dbMySQL()
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return fmt.Errorf("open mysql err: %w", err)
	}
	mysqlDB = db
	return nil
}

func initRedis() error {
	opt, err := redis.ParseURL(dbRedis())
	if err != nil {
		return fmt.Errorf("parse redis url err: %w", err)
	}

	redisCli = redis.NewClient(opt)
	return nil
}
