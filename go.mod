module github.com/JaloMu/libs

go 1.18

require (
	github.com/JaloMu/tools v0.0.3
	github.com/gin-gonic/gin v1.7.7
	github.com/spf13/viper v1.10.1
	github.com/tidwall/gjson v1.12.1
	go.uber.org/zap v1.20.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

require (
	github.com/BurntSushi/toml v0.4.1 // indirect
	gorm.io/driver/mysql v1.2.3
	gorm.io/gorm v1.22.5
	moul.io/zapgorm2 v1.1.1
)

replace github.com/JaloMu/tools v0.0.3 => ../tools
