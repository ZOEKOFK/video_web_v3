package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	driverName := "mysql"
	host := "localhost" //linux这里写的是host变为db
	port := "3306"
	dataname := "video_web"
	username := "root"
	password := "zhangqi20060212"
	charset := "utf8"
	s := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		username, password, host, port, dataname, charset)
	db, err := gorm.Open(driverName, s)
	if err != nil {
		panic("failed to connect database" + err.Error())
	} else {
		fmt.Println("数据库连接成功(●'◡'●)")
	}
	DB = db
	return db, err
}

func GetDB() *gorm.DB {
	return DB
}
