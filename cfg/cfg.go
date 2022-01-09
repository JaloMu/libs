package cfg

import (
	"errors"
	"path"
	"strings"

	"github.com/spf13/viper"
)

var (
	ErrorConfigSuffix   = errors.New("config suffix error")
	ErrorConfigFileLoad = errors.New("config file load error")
	ErrorConfigParse    = errors.New("config file parse error")
)

type config struct {
	fileSlice []string
	Conf      interface{}
}

func NewConfig(filepath string, confStruct interface{}) (c *config, err error) {
	configDir, configFilename, configSuffix, err := parseFilePath(filepath)
	c = &config{
		fileSlice: []string{configDir, configFilename, configSuffix},
		Conf:      confStruct,
	}
	return c, nil
}

func parseFilePath(filepath string) (configDir, configFilename, configSuffix string, err error) {
	// 获取文件后缀
	var suffix = path.Ext(filepath)
	configDir = path.Dir(filepath)
	configFilename = strings.TrimSuffix(path.Base(filepath), suffix)
	switch suffix {
	case ".toml":
		return configDir, configFilename, "toml", err
	case ".json":
		return configDir, configFilename, "json", err
	default:
		return "", "", "", ErrorConfigSuffix
	}
}

func (c *config) ParseConfig() (err error) {
	viper.SetConfigType(c.fileSlice[2])
	viper.SetConfigName(c.fileSlice[1])
	viper.AddConfigPath(c.fileSlice[0])
	if err = viper.ReadInConfig(); err != nil {
		return ErrorConfigFileLoad
	}

	if err = viper.Unmarshal(c.Conf); err != nil {
		return ErrorConfigParse
	}
	return nil
}
