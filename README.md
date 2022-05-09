# model-filter
基于 gin 和 gorm 的 model filter

### 1. 初始化

a. 创建 ModelFilter：

`f := filters.New(User{})`

b. 从 gin.Context 初始化一个 ModelFilter：

`f := filters.New(User{}, ctx)`


### 2. model 设置 Tag

创建 ModelFilter 需要传入一个 model 对象，model 中需要给相应的字段加上如下 tag (可选):

`filter:"name:xxx;order;search;match"`

其中:
  - "name:xxx": 可选，xxx 是 db 中的字段名称，若不设置则默认为字段名的 snake_case
  - "order": 可选，表示允许以该字段排序
  - "search": 可选，表示允许以该字段搜索
  - "match": 可选，表示允许以该字段匹配
  
 

### 3. 调用

函数：

`func (f *ModelFilter) Query(db *gorm.DB) *gorm.DB`

示例：

`mf.Query(db).Find(&Foo).Error`


### 4. 功能

- 排序：`?_order=xxx`, 前加 "-" 表示降序
- 分页：`?_limit=10&_offset=10`，不设置_limit返回所有记录，_offset默认为0
- 按字段搜索：`?_search_fields=xxx,yyy&_search=vvv`, 多个字段用","分隔
- 按字段过滤：`?name=xxx&age=12`
- 自定义查询：对于复杂的查询逻辑，可以自定义查询语句

其中，默认的功能性 key 为 _limit/_offset/_order/_search_fields/_search，可通过 SetGlobalConfig 函数修改默认 key
非功能性 key 将被视为按字段匹配
