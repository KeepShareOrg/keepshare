// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/alecthomas/participle/v2"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

// listSharedLinks query the current user shared links by filter
func listSharedLinks(c *gin.Context) {
	userID := c.GetString(constant.UserID)

	searchValue := c.Query("search")
	filter := c.Query("filter")

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 10
	}

	pageIndex, err := strconv.Atoi(c.Query("page_index"))
	if err != nil {
		pageIndex = 1
	}

	var filters []Query
	if filter != "" {
		if err := json.Unmarshal([]byte(filter), &filters); err != nil {
			c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params"))
			return
		}
	}

	if searchValue != "" {
		dslFilters, err := parseQueryDSL(searchValue)

		if err != nil {
			c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params: "+err.Error()))
			return
		}

		// merge filters
		for _, filter := range dslFilters {
			filters = append(filters, filter)
		}
	}

	ctx := c.Request.Context()
	ret, total, err := conditionQuery(ctx, filters, userID, pageIndex, limit)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	if len(ret) > 0 {
		hashes := make([]string, 0, len(ret))
		for _, v := range ret {
			hashes = append(hashes, v.OriginalLinkHash)
		}
		t := query.Blacklist
		t.WithContext(ctx).Where(t.UserID.Eq(userID), t.OriginalLinkHash.In(hashes...))
	}

	log.WithField("shared_records", ret).Debugf("condition query result")

	c.JSON(http.StatusOK, Map{
		"total":     total,
		"page_size": len(ret),
		"list":      ret,
	})
}

// querySharedLinkInfo query shared link current status and this shared link's info
func querySharedLinkInfo(c *gin.Context) {
	id := c.Query("id")
	autoID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params"))
		return
	}

	conditions := []gen.Condition{
		query.SharedLink.AutoID.Eq(int64(autoID)),
	}
	res, err := query.SharedLink.Where(conditions...).Take()
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

// batchQuerySharedLinksInfo batch query shared link current status and this shared link's info
func batchQuerySharedLinksInfo(c *gin.Context) {
	type request struct {
		Links []string `json:"links"`
	}

	req := new(request)
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params"))
		return
	}

	if len(req.Links) == 0 {
		c.JSON(http.StatusOK, gin.H{"list": []interface{}{}})
		return
	}

	hashes := make([]string, 0)
	for _, link := range req.Links {
		hashes = append(hashes, lk.Hash(link))
	}

	userID := c.GetString(constant.UserID)
	conditions := []gen.Condition{
		query.SharedLink.OriginalLinkHash.In(hashes...),
		query.SharedLink.UserID.Eq(userID),
	}

	ret, err := query.SharedLink.Where(conditions...).Find()
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list": ret,
	})
}

type QueryOperator string

const (
	OpBetween        QueryOperator = "between"
	OpMatch          QueryOperator = ":"
	OpEquals         QueryOperator = "="
	OpNotEquals      QueryOperator = "!="
	OpGreaterThan    QueryOperator = ">"
	OpGreaterOrEqual QueryOperator = ">="
	OpLessThan       QueryOperator = "<"
	OpLessOrEqual    QueryOperator = "<="
)

type SupportQueryKey string

type Query struct {
	Key      SupportQueryKey `json:"key"`
	Operator QueryOperator   `json:"operator"`
	Value    any             `json:"value"`
}

var types = make(map[SupportQueryKey]any)

func initQueryTypes() {
	table := query.SharedLink
	for _, f := range []field.Expr{
		table.Title,
		table.OriginalLink, // TODO query by hash
		table.HostSharedLink,
		table.CreatedAt,
		table.CreatedBy,
		table.State,
		table.Stored,
		table.Visitor,
		table.Size,
		table.LastVisitedAt,
	} {
		types[SupportQueryKey(f.ColumnName().String())] = f
	}
}

// conditionQuery shared link query method, support query by filter condition
func conditionQuery(ctx context.Context, filters []Query, userID string, pageIndex, limit int) ([]*model.SharedLink, int64, error) {
	conditions := make([]gen.Condition, 0)

	for _, f := range filters {
		switch v := types[f.Key].(type) {
		case field.Int32:
			if f.Operator == OpBetween {
				val, ok := f.Value.([]interface{})
				if ok && len(val) == 2 {
					v1, ok1 := val[0].(float64)
					v2, ok2 := val[1].(float64)
					if ok1 && ok2 {
						conditions = append(conditions, v.Between(int32(v1), int32(v2)))
					}
				} else {
					// []float64 from parse dsl
					val, ok := f.Value.([]float64)
					if ok && len(val) == 2 {
						conditions = append(conditions, v.Between(int32(val[0]), int32(val[1])))
					}
				}
			} else if val, ok := f.Value.(float64); ok {
				switch f.Operator {
				case OpEquals:
					conditions = append(conditions, v.Eq(int32(val)))
				case OpNotEquals:
					conditions = append(conditions, v.Neq(int32(val)))
				case OpGreaterThan:
					conditions = append(conditions, v.Gt(int32(val)))
				case OpGreaterOrEqual:
					conditions = append(conditions, v.Gte(int32(val)))
				case OpLessThan:
					conditions = append(conditions, v.Lt(int32(val)))
				case OpLessOrEqual:
					conditions = append(conditions, v.Lte(int32(val)))
				}
			}
		case field.Int64:
			if f.Operator == OpBetween {
				val, ok := f.Value.([]interface{})
				if ok && len(val) == 2 {
					v1, ok1 := val[0].(float64)
					v2, ok2 := val[1].(float64)
					if ok1 && ok2 {
						conditions = append(conditions, v.Between(int64(v1), int64(v2)))
					}
				} else {
					// []float64 from parse dsl
					val, ok := f.Value.([]float64)
					if ok && len(val) == 2 {
						conditions = append(conditions, v.Between(int64(val[0]), int64(val[1])))
					}
				}
			} else if val, ok := f.Value.(float64); ok {
				switch f.Operator {
				case OpEquals:
					conditions = append(conditions, v.Eq(int64(val)))
				case OpNotEquals:
					conditions = append(conditions, v.Neq(int64(val)))
				case OpGreaterThan:
					conditions = append(conditions, v.Gt(int64(val)))
				case OpGreaterOrEqual:
					conditions = append(conditions, v.Gte(int64(val)))
				case OpLessThan:
					conditions = append(conditions, v.Lt(int64(val)))
				case OpLessOrEqual:
					conditions = append(conditions, v.Lte(int64(val)))
				}
			}
		case field.Time:
			if f.Operator == OpBetween {
				val, ok := f.Value.([]interface{})
				if ok && len(val) == 2 {
					t1, err1 := time.ParseInLocation("2006-01-02 15:04:05", val[0].(string), time.Local)
					t2, err2 := time.ParseInLocation("2006-01-02 15:04:05", val[1].(string), time.Local)
					if err1 == nil && err2 == nil {
						conditions = append(conditions, v.Between(t1, t2))
					}
				}
			} else if val, ok := f.Value.(string); ok {
				if t, err := time.ParseInLocation("2006-01-02 15:04:05", val, time.Local); err == nil {
					switch f.Operator {
					case OpEquals:
						conditions = append(conditions, v.Eq(t))
					case OpNotEquals:
						conditions = append(conditions, v.Neq(t))
					case OpGreaterThan:
						conditions = append(conditions, v.Gt(t))
					case OpGreaterOrEqual:
						conditions = append(conditions, v.Gte(t))
					case OpLessThan:
						conditions = append(conditions, v.Lt(t))
					case OpLessOrEqual:
						conditions = append(conditions, v.Lte(t))
					}
				}
			}
		case field.String:
			if val, ok := f.Value.(string); ok {
				switch f.Operator {
				case OpMatch:
					conditions = append(conditions, v.Like("%"+val+"%"))
				case OpEquals:
					conditions = append(conditions, v.Eq(val))
				case OpNotEquals:
					conditions = append(conditions, v.Neq(val))
				case OpGreaterThan:
					conditions = append(conditions, v.Gt(val))
				case OpGreaterOrEqual:
					conditions = append(conditions, v.Gte(val))
				case OpLessThan:
					conditions = append(conditions, v.Lt(val))
				case OpLessOrEqual:
					conditions = append(conditions, v.Lte(val))
				}
			}
		}
	}

	conditions = append(conditions, query.SharedLink.UserID.Eq(userID))

	// Only the first query returns the real total num
	total := int64(0)
	if pageIndex == 1 {
		num, err := query.SharedLink.WithContext(ctx).Where(conditions...).Count()
		if err != nil {
			return nil, 0, err
		}
		total = num
	}

	ret, err := query.SharedLink.WithContext(ctx).Where(conditions...).Order(query.SharedLink.AutoID.Desc()).Offset((pageIndex - 1) * limit).Limit(limit).Find()
	if err != nil {
		return nil, 0, err
	}

	return ret, total, nil
}

type Condition struct {
	Key   string       `parser:"@Ident"`
	Op    string       `parser:"@('!'? '=' | '<' '='? | '>' '='? | ':')"`
	Value *FilterValue `parser:"@@"`
}

type FilterValue struct {
	Bool          *bool     `parser:"@('true' | 'false')"`
	String        *string   `parser:"| @String"`
	Number        *float64  `parser:"| @(Int | Float)"`
	BetweenString []string  `parser:"| '[' @String (',' @String)* ']'"`
	BetweenNumber []float64 `parser:"| '[' @(Int | Float) (',' @(Int | Float))* ']'"`
}

// keepShare query dsl parser
var parser = participle.MustBuild[struct {
	Conditions []Condition `parser:"@@*"`
}](participle.Unquote("String"))

var GBReg = regexp.MustCompile(`(?i)([\d.]+)GB$`)

// transferSupportUnit transfer support unit string to float64
func transferSupportUnit(input string) (float64, error) {
	ret := GBReg.FindAllStringSubmatch(input, -1)
	if len(ret) > 0 {
		unit, err := strconv.ParseFloat(ret[0][1], 64)
		if err != nil {
			return 0, err
		}
		// per GB equal 1024 * 1024 * 1024 bytes
		return unit * 1024 * 1024 * 1024, nil
	}

	return 0, errors.New("unsupported unit")
}

// parseQueryDSL parse keepShare dsl string to []Query
func parseQueryDSL(queryString string) ([]Query, error) {
	ret, err := parser.ParseString("", queryString)
	if err != nil {
		return nil, err
	}

	filters := make([]Query, 0)

	for _, f := range ret.Conditions {
		filter := Query{
			Key:      SupportQueryKey(string(f.Key)),
			Operator: QueryOperator(string(f.Op)),
		}

		v := f.Value
		if v.Bool != nil {
			filter.Value = *v.Bool
		} else if v.Number != nil {
			filter.Value = *v.Number
		} else if v.String != nil {
			if unit, err := transferSupportUnit(*v.String); err != nil {
				filter.Value = *v.String
			} else {
				filter.Value = unit
			}
		} else if v.BetweenNumber != nil {
			filter.Value = v.BetweenNumber
		} else if v.BetweenString != nil {
			filter.Value = v.BetweenString
			tempFloats := make([]float64, 0)
			for _, str := range v.BetweenString {
				if unit, err := transferSupportUnit(str); err != nil {
					filter.Value = v.BetweenString
					break
				} else {
					tempFloats = append(tempFloats, unit)
				}
			}
			if len(tempFloats) == 2 {
				filter.Value = tempFloats
			}
		}

		filters = append(filters, filter)
	}

	return filters, nil
}

func deleteSharedLinks(c *gin.Context) {
	var req struct {
		Links []string `json:"links"`
		Host  string   `json:"host"`
	}
	var resp struct {
		RowsAffected int64    `json:"rows_affected"`
		ErrorLinks   []string `json:"error_links"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params"))
		return
	}

	if len(req.Links) == 0 {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "links_is_empty"))
		return
	}

	hostName := util.FirstNotEmpty(req.Host, config.DefaultHost())
	host := hosts.Get(hostName)
	if host == nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_host", i18n.WithDataMap("host", hostName)))
		return
	}

	ctx := c.Request.Context()
	userID := c.GetString(constant.UserID)
	original, invalid := getOriginalLinks(req.Links)
	resp.ErrorLinks = invalid

	rows, err := deleteSharedLinksWithHost(ctx, userID, host, original)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}
	resp.RowsAffected = rows
	c.JSON(http.StatusOK, resp)
}

func deleteSharedLinksWithHost(ctx context.Context, userID string, host hosts.Host, original []string) (rowsAffected int64, err error) {
	if len(original) == 0 {
		return 0, nil
	}

	if err := host.Delete(ctx, userID, original); err != nil {
		return 0, err
	}

	hashes := make([]string, 0, len(original))
	for _, link := range original {
		hashes = append(hashes, lk.Hash(link))
	}

	conditions := []gen.Condition{
		query.SharedLink.OriginalLinkHash.In(hashes...),
		query.SharedLink.UserID.Eq(userID),
	}

	ret, err := query.SharedLink.WithContext(ctx).Where(conditions...).Delete()
	if err != nil {
		return 0, err
	}

	return ret.RowsAffected, nil
}
