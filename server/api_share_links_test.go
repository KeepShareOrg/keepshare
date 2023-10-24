// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"encoding/json"
	"testing"
)

func TestParserDSL(t *testing.T) {
	searchStr := `
username = "admin";
password != "123456";
age < 20;
hello > 100;
n1 >= 30;
n2 <= 40;
ok = true;
height between [160, 168];
xx between ["x", "y"];
yy between [123.456, 789.123];
name match "this is string test\"ok\"";
`
	ret, err := parseQueryDSL(searchStr)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	if bs, err := json.MarshalIndent(ret, "", "  "); err == nil {
		t.Logf("ret: %v", string(bs))
	}
}
