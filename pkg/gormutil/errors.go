// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gormutil

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// IsDuplicateError checks if the given error is duplicate entry error.
func IsDuplicateError(err error) bool {
	if err != nil {
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return true
		}
	}
	return false
}

// IsNotFoundError checks if the given error is record not found error.
func IsNotFoundError(err error) bool {
	return err != nil && errors.Is(err, gorm.ErrRecordNotFound)
}
