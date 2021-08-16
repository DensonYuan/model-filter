package filters

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
	"strconv"
	"strings"
)

type ModelFilter struct {
	model             interface{}
	orderBy           string
	searchFields      string
	searchValue       string
	mapFieldMatch     map[string]interface{}
	queryList         []string
	argsList          [][]interface{}
	limit             int
	offset            int
	fields            string
	preloadColumn     string
	preloadConditions []interface{}

	// 功能性字段集合 (排序/搜索/匹配)
	allowOrderFields  map[string]struct{}
	allowMatchFields  map[string]struct{}
	allowSearchFields map[string]struct{}
}

//////////////////////////////////////////////////////////////////////////////////////

const (
	defaultLimitKey        = "_limit"
	defaultOffsetKey       = "_offset"
	defaultOrderKey        = "_order"
	defaultSearchFieldsKey = "_search_fields"
	defaultSearchValueKey  = "_search"
	defaultFieldsKey       = "_fields"
)

var globalConfig = &Config{
	LimitKey:        defaultLimitKey,
	OffsetKey:       defaultOffsetKey,
	OrderKey:        defaultOrderKey,
	SearchFieldsKey: defaultSearchFieldsKey,
	SearchValueKey:  defaultSearchValueKey,
	FieldsKey:       defaultFieldsKey,
}

// 设置 url 中的功能性字段
type Config struct {
	LimitKey        string
	OffsetKey       string
	OrderKey        string
	SearchFieldsKey string
	SearchValueKey  string
	FieldsKey       string
}

func isFunctionalKey(key string) bool {
	return key == globalConfig.LimitKey || key == globalConfig.OffsetKey || key == globalConfig.OrderKey ||
		key == globalConfig.SearchFieldsKey || key == globalConfig.SearchValueKey || key == globalConfig.FieldsKey
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) initFromGinContext(c *gin.Context) {
	f.limit, _ = strconv.Atoi(c.DefaultQuery(globalConfig.LimitKey, "-1"))
	f.offset, _ = strconv.Atoi(c.DefaultQuery(globalConfig.OffsetKey, "0"))
	f.orderBy = c.DefaultQuery(globalConfig.OrderKey, "")
	f.searchFields = c.DefaultQuery(globalConfig.SearchFieldsKey, "")
	f.searchValue = c.DefaultQuery(globalConfig.SearchValueKey, "")
	f.fields = c.DefaultQuery(globalConfig.FieldsKey, "")

	m := (map[string][]string)(c.Request.URL.Query())
	for k, v := range m {
		if !isFunctionalKey(k) && len(v) > 0 {
			f.Match(k, v[0])
		}
	}
}

func (f *ModelFilter) initFunctionalFields() {
	f.allowOrderFields = make(map[string]struct{})
	f.allowMatchFields = make(map[string]struct{})
	f.allowSearchFields = make(map[string]struct{})

	modelType := reflect.TypeOf(f.model)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		fieldName := snakeCase(field.Name)
		tags := strings.Split(field.Tag.Get("filter"), ";")
		if len(tags) > 0 && strings.HasPrefix(tags[0], "name:") {
			fieldName = tags[0][5:]
		}
		for _, t := range tags {
			if t == "order" {
				f.allowOrderFields[fieldName] = struct{}{}
			} else if t == "search" {
				f.allowSearchFields[fieldName] = struct{}{}
			} else if t == "match" {
				f.allowMatchFields[fieldName] = struct{}{}
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) orderHandler(db *gorm.DB) *gorm.DB {
	if f.orderBy != "" {
		obc := clause.OrderByColumn{Column: clause.Column{Name: f.orderBy}}
		if strings.HasPrefix(f.orderBy, "-") {
			obc.Desc = true
			obc.Column.Name = f.orderBy[1:]
		}
		if _, ok := f.allowOrderFields[obc.Column.Name]; ok {
			db = db.Order(obc)
		}
	}
	return db
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) paginationHandler(db *gorm.DB) *gorm.DB {
	db = db.Limit(f.limit)
	db = db.Offset(f.offset)
	return db
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) searchHandler(db *gorm.DB) *gorm.DB {
	if f.searchValue == "" {
		return db
	}
	var clauses []string
	format := "`%s` LIKE '%%%s%%'"
	if f.searchFields != "" {
		for _, field := range strings.Split(f.searchFields, ",") {
			if _, ok := f.allowSearchFields[field]; ok && field != "" {
				clauses = append(clauses, fmt.Sprintf(format, field, f.searchValue))
			}
		}
	} else {
		for field := range f.allowSearchFields {
			clauses = append(clauses, fmt.Sprintf(format, field, f.searchValue))
		}
	}
	db = db.Where(strings.Join(clauses, " OR "))
	return db
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) matchHandler(db *gorm.DB) *gorm.DB {
	for k, v := range f.mapFieldMatch {
		if _, ok := f.allowMatchFields[k]; ok {
			db = db.Where(fmt.Sprintf("`%s` = ?", k), v)
		}
	}
	return db
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) clauseHandler(db *gorm.DB) *gorm.DB {
	for i := range f.queryList {
		db = db.Where(f.queryList[i], f.argsList[i]...)
	}
	return db
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) selectHandler(db *gorm.DB) *gorm.DB {
	if f.fields != "" {
		return db.Select(strings.Split(f.fields, ","))
	}
	return db
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) preloadHandler(db *gorm.DB) *gorm.DB {
	if f.preloadColumn != "" {
		return db.Preload(f.preloadColumn, f.preloadConditions...)
	}
	return db
}
