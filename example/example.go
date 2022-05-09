package main

import (
	"fmt"
	"github.com/DensonYuan/model-filter"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

var DB *gorm.DB

type User struct {
	Name  string `json:"name,omitempty" filter:"order;search;match"`
	Age   int    `json:"age,omitempty" filter:"order;match"`
	Email string `json:"email,omitempty" filter:"search;match"`
}

func (*User) TableName() string {
	return "user"
}

func init() {
	ptn := "%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=%s"
	dsn := fmt.Sprintf(ptn, "root", "", "localhost", 3306, "test", "Asia%2FShanghai")
	DB, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{AllowGlobalUpdate: true})
	DB = DB.Debug()

	// 可以通过 SetGlobalConfig 设置默认的功能性 key
	//filters.SetGlobalConfig(&filters.Config{
	//	LimitKey:        "limit",
	//	OffsetKey:       "offset",
	//	OrderKey:        "order",
	//	SearchFieldsKey: "search_fields",
	//	SearchValueKey:  "search_value",
	//	FieldsKey:       "fields",
	//})
}

func migrate() {
	d := DB.Set("gorm:table_options", "DEFAULT CHARSET=utf8mb4")
	d.AutoMigrate(&User{})
}

func main() {
	//migrate()
	StartAPIServer()
}

func StartAPIServer() {
	r := gin.Default()
	r.GET("/api/list/", ListHandler)
	r.Run(":80")
}

func ListHandler(c *gin.Context) {
	filter := filters.New(User{}, c)

	// 删除
	//_ = filter.Limit(1).Delete(DB)

	// 手动指定返回字段
	//filter.Select("name,age")

	//手动指定匹配字段
	//filter.Match("name", "tom")

	// 手动指定复杂查询语句
	//filter.Where("name = ? AND age > ?", "tom", 12)

	// 支持 where 链式调用
	//filter.Where("email LIKE '%@%'")

	// 计数
	//cnt, _ := filter.Count(DB)
	//fmt.Println(cnt)

	var users []User
	if err := filter.Query(DB).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, users)
}
