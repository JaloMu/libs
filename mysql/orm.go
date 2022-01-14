package mysql

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Options struct {
	Address       string        `json:"address"`
	SlowThreshold time.Duration `json:"slow_threshold"`
	MaxOpen       int           `json:"max_open"`
	MaxIdle       int           `json:"max_idle"`
	L             *zap.Logger   `json:"l"`
	Namespace     string        `json:"namespace"`
}

func InitOrmMysql(d Options) (engine *gorm.DB, err error) {
	var l = New(d.L.Named(d.Namespace), "trace")
	l.SetAsDefault()
	var gc = gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		Logger:      l,
		PrepareStmt: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}
	engine, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       d.Address, // DSN data source name
		DefaultStringSize:         256,       // string 类型字段的默认长度
		DisableDatetimePrecision:  true,      // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,      // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,      // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,
	}), &gc)
	sqlDB, err := engine.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(d.MaxIdle)
	sqlDB.SetMaxOpenConns(d.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Minute)
	return
}
