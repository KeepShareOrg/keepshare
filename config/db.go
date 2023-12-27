// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"

	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	gormDB   *gorm.DB
	redisCli *redis.Client
)

// MySQL returns the mysql client instance.
func MySQL() *gorm.DB {
	if gormDB == nil {
		log.Fatal("mysql has not been initialized")
	}
	return gormDB
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
	gormDB = db

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("init sql db err: %w", err)
	}

	if n := viper.GetInt("db_mysql_max_open_conns"); n > 0 {
		sqlDB.SetMaxOpenConns(n)
		log.Debugf("mysql set max_open_conns: %d", n)
	}
	if n := viper.GetInt("db_mysql_max_idle_conns"); n > 0 {
		sqlDB.SetMaxIdleConns(n)
		log.Debugf("mysql set max_idle_conns: %d", n)
	}
	if d := viper.GetDuration("db_mysql_max_idle_time"); d > 0 {
		sqlDB.SetConnMaxIdleTime(d)
		log.Debugf("mysql set max_idle_time: %s", d)
	}
	if d := viper.GetDuration("db_mysql_max_life_time"); d > 0 {
		sqlDB.SetConnMaxLifetime(d)
		log.Debugf("mysql set max_life_time: %s", d)
	}

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
