package response

import (
	"net/http"

	"github.com/JaloMu/libs/I18n"

	"github.com/gin-gonic/gin"
)

type response struct {
	Status int64       `json:"status"`
	Type   string      `json:"type,omitempty"`
	Msg    string      `json:"msg,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

func Json(ctx *gin.Context, key string, data ...interface{}) {
	var r response
	var language = ctx.GetString("language")
	if language != "zh" && language != "en" {
		language = "en"
	}
	result := I18n.Get(language, key)
	r = response{
		Status: result.Map()["code"].Int(),
		Type:   result.Map()["type"].String(),
		Msg:    result.Map()["msg"].String(),
		Data:   data,
	}
	if len(data) == 0 {
		r.Data = nil
	}
	if len(data) == 1 {
		r.Data = data[0]
	}
	ctx.JSON(http.StatusOK, r)
	ctx.Abort()
}
