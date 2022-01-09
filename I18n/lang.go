package I18n

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/JaloMu/tools/utils"

	"github.com/tidwall/gjson"
)

type Locale struct {
	zh string
	en string
}

var l *Locale

func New(dir string) (err error) {
	var zhLocaleFile = path.Join(dir, "zh.json")
	var enLocaleFile = path.Join(dir, "en.json")
	if !utils.IsExist(zhLocaleFile) {
		return errors.New("locale file is not exist")
	}
	if !utils.IsFile(zhLocaleFile) {
		return errors.New("locale file is not file")
	}
	if !utils.IsExist(enLocaleFile) {
		return errors.New("locale file is not exist")
	}
	if !utils.IsFile(enLocaleFile) {
		return errors.New("locale file is not file")
	}
	zhFile, err := os.Open(zhLocaleFile)
	if err != nil {
		return err
	}
	defer func(zhFile *os.File) {
		_ = zhFile.Close()
	}(zhFile)
	enFile, err := os.Open(enLocaleFile)
	if err != nil {
		return
	}
	defer func(zhFile *os.File) {
		_ = zhFile.Close()
	}(zhFile)
	zhByte, err := ioutil.ReadAll(zhFile)
	if err != nil {
		return
	}
	enByte, err := ioutil.ReadAll(enFile)
	if err != nil {
		return
	}
	var data = make(map[string]interface{})
	err = json.Unmarshal(zhByte, &data)
	if err != nil {
		return
	}
	err = json.Unmarshal(enByte, &data)
	if err != nil {
		return
	}
	l = &Locale{
		zh: string(zhByte),
		en: string(enByte),
	}
	return
}

func Get(locale, key string) (result gjson.Result) {
	switch locale {
	case "zh":
		return gjson.Get(l.zh, key)
	case "en":
		return gjson.Get(l.en, key)
	default:
		return gjson.Result{}
	}
}
