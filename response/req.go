package response

import (
	"net/http"

	"github.com/JaloMu/libs/I18n"
	"github.com/JaloMu/libs/trace"

	"github.com/gin-gonic/gin"
)

type response struct {
	Status  int64       `json:"status"`
	Type    string      `json:"type,omitempty"`
	Msg     string      `json:"msg,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	TraceId string      `json:"trace_id,omitempty"`
	SpanId  string      `json:"span_id,omitempty"`
	CSpanId string      `json:"c_span_id,omitempty"`
}

func Json(ctx *gin.Context, traceKeys, key string, data ...interface{}) {
	var r response
	var language = ctx.GetString("language")
	if language != "zh" && language != "en" {
		language = "en"
	}
	if len(traceKeys) == 0 {
		traceKeys = "trace"
	}
	tr, _ := ctx.Get(traceKeys)
	traceContext, _ := tr.(*trace.TraceContext)
	result := I18n.Get(language, key)
	r = response{
		Status:  result.Map()["code"].Int(),
		Type:    result.Map()["type"].String(),
		Msg:     result.Map()["msg"].String(),
		Data:    data,
		TraceId: traceContext.TraceId,
		SpanId:  traceContext.CSpanId,
		CSpanId: traceContext.SpanId,
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
