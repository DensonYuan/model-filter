package filters

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 通过 gin.Context 初始化 ModelFilter
func InitModelFilter(c *gin.Context, model interface{}) *ModelFilter {
	mf := &ModelFilter{model: model}
	mf.initFromGinContext(c)
	mf.initFunctionalFields()
	return mf
}

// 创建 ModelFilter，传入 model 对象
func NewModelFilter(model interface{}) *ModelFilter {
	mf := &ModelFilter{model: model}
	mf.initFunctionalFields()
	return mf
}

// 设置全局配置
func SetGlobalConfig(config *Config) {
	globalConfig = config
	if globalConfig.LimitKey == "" {
		globalConfig.LimitKey = defaultLimitKey
	}
	if globalConfig.OffsetKey == "" {
		globalConfig.OffsetKey = defaultOffsetKey
	}
	if globalConfig.OrderKey == "" {
		globalConfig.OrderKey = defaultOrderKey
	}
	if globalConfig.SearchFieldsKey == "" {
		globalConfig.SearchFieldsKey = defaultSearchFieldsKey
	}
	if globalConfig.SearchValueKey == "" {
		globalConfig.SearchValueKey = defaultSearchValueKey
	}
	if globalConfig.FieldsKey == "" {
		globalConfig.FieldsKey = defaultFieldsKey
	}
}

// 获取结果集合
func (f *ModelFilter) Query(db *gorm.DB) *gorm.DB {
	db = db.Model(f.model)
	db = f.orderHandler(db)
	db = f.searchHandler(db)
	db = f.matchHandler(db)
	db = f.clauseHandler(db)
	db = f.paginationHandler(db)
	db = f.selectHandler(db)
	db = f.preloadHandler(db)
	return db
}

// 获取计数结果
func (f *ModelFilter) Count(db *gorm.DB) (cnt int64, err error) {
	err = f.Query(db).Limit(-1).Offset(-1).Count(&cnt).Error
	return
}

// 直接删除匹配的记录
func (f *ModelFilter) Delete(db *gorm.DB) (err error) {
	err = f.Query(db).Delete(f.model).Error
	return
}

// 设置查询字段
func (f *ModelFilter) Select(fields string) *ModelFilter {
	f.fields = fields
	return f
}

// 设置 Where 查询条件
func (f *ModelFilter) Where(query string, args ...interface{}) *ModelFilter {
	f.queryList = append(f.queryList, query)
	f.argsList = append(f.argsList, args)
	return f
}

// 设置字段匹配条件
func (f *ModelFilter) Match(field string, value interface{}) *ModelFilter {
	if f.mapFieldMatch == nil {
		f.mapFieldMatch = make(map[string]interface{})
	}
	f.mapFieldMatch[field] = value
	return f
}

// 设置排序字段
func (f *ModelFilter) Order(value string) *ModelFilter {
	f.orderBy = value
	return f
}

// 设置分页大小
func (f *ModelFilter) Limit(limit int) *ModelFilter {
	f.limit = limit
	return f
}

// 设置分页偏移
func (f *ModelFilter) Offset(offset int) *ModelFilter {
	f.offset = offset
	return f
}

// 设置搜索字段及值
func (f *ModelFilter) Search(fields string, value string) *ModelFilter {
	f.searchFields = fields
	f.searchValue = value
	return f
}

// 设置预加载条件
func (f *ModelFilter) Preload(column string, conditions ...interface{}) *ModelFilter {
	f.preloadColumn = column
	f.preloadConditions = conditions
	return f
}
