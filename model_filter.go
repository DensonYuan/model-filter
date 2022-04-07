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

// ModelFilter exported model filter
type ModelFilter struct {
	model        interface{}
	orderBy      string
	limit        int
	offset       int
	selectFields string
	searchFields string
	searchValue  string
	queries      []queryPair
	joins        []joinPair
	matches      map[string]interface{}
	preloads     map[string][]interface{}

	// 功能性字段集合 (排序/搜索/匹配)
	canOrder  map[string]bool
	canMatch  map[string]bool
	canSearch map[string]bool
}

type queryPair struct {
	Query interface{}
	Args  []interface{}
}

type joinPair struct {
	Query string
	Args  []interface{}
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

// Config url 中的功能性字段设置
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
	f.selectFields = c.DefaultQuery(globalConfig.FieldsKey, "")
	f.searchFields = c.DefaultQuery(globalConfig.SearchFieldsKey, "")
	f.searchValue = c.DefaultQuery(globalConfig.SearchValueKey, "")

	m := (map[string][]string)(c.Request.URL.Query())
	for k, v := range m {
		if !isFunctionalKey(k) && len(v) > 0 && v[0] != "" {
			f.Match(k, v[0])
		}
	}
}

func (f *ModelFilter) initFunctionalFields() {
	// TODO: 根据 model 加对应缓存，避免每次初始化，benchmark 测试下性能
	f.canOrder = make(map[string]bool)
	f.canMatch = make(map[string]bool)
	f.canSearch = make(map[string]bool)

	modelType := reflect.TypeOf(f.model)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		fieldName := snakeCase(field.Name)
		tags := strings.Split(field.Tag.Get("filter"), ";")
		if len(tags) > 0 && strings.HasPrefix(tags[0], "name:") {
			fieldName = tags[0][5:]
		}
		// TODO: 增加 inset 功能
		for _, t := range tags {
			if t == "order" {
				f.canOrder[fieldName] = true
			} else if t == "search" {
				f.canSearch[fieldName] = true
			} else if t == "match" {
				f.canMatch[fieldName] = true
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////

func (f *ModelFilter) joinHandler(db *gorm.DB) *gorm.DB {
	for _, p := range f.joins {
		db = db.Joins(p.Query, p.Args...)
	}
	return db
}

func (f *ModelFilter) orderHandler(db *gorm.DB) *gorm.DB {
	if f.orderBy != "" {
		obc := clause.OrderByColumn{Column: clause.Column{Name: f.orderBy}}
		if strings.HasPrefix(f.orderBy, "-") {
			obc.Desc = true
			obc.Column.Name = f.orderBy[1:]
		}
		if f.canOrder[obc.Column.Name] {
			db = db.Order(obc)
		}
	}
	return db
}

func (f *ModelFilter) paginationHandler(db *gorm.DB) *gorm.DB {
	db = db.Limit(f.limit)
	db = db.Offset(f.offset)
	return db
}

func (f *ModelFilter) searchHandler(db *gorm.DB) *gorm.DB {
	if f.searchValue == "" {
		return db
	}
	var clauses []string
	format := "`%s` LIKE '%%%s%%'"
	if f.searchFields != "" {
		for _, field := range strings.Split(f.searchFields, ",") {
			if field != "" && f.canSearch[field] {
				clauses = append(clauses, fmt.Sprintf(format, field, f.searchValue))
			}
		}
	} else {
		for field := range f.canSearch {
			clauses = append(clauses, fmt.Sprintf(format, field, f.searchValue))
		}
	}
	db = db.Where(strings.Join(clauses, " OR "))
	return db
}

func (f *ModelFilter) matchHandler(db *gorm.DB) *gorm.DB {
	for k, v := range f.matches {
		if f.canMatch[k] {
			var vs []string
			if s, ook := v.(string); ook {
				vs = strings.Split(s, ",")

			}
			if len(vs) > 1 {
				db = db.Where(fmt.Sprintf("`%s` in (?)", k), vs)
			} else {
				db = db.Where(fmt.Sprintf("`%s` = ?", k), v)
			}
		}
	}
	return db
}

func (f *ModelFilter) clauseHandler(db *gorm.DB) *gorm.DB {
	for _, p := range f.queries {
		db = db.Where(p.Query, p.Args...)
	}
	return db
}

func (f *ModelFilter) selectHandler(db *gorm.DB) *gorm.DB {
	if f.selectFields != "" {
		return db.Select(strings.Split(f.selectFields, ","))
	}
	return db
}

func (f *ModelFilter) preloadHandler(db *gorm.DB) *gorm.DB {
	for query, args := range f.preloads {
		db = db.Preload(query, args...)
	}
	return db
}
