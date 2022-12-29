package main

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func readIDParam(r *http.Request) (int64, error) {
	//解析请求参数，将查询的id转为int类型，因为可能后期数据量比较大，所以用int64类型
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		//通过http.Error(w, err.Error(), http.StatusBadRequest)将这里生成的错误信息返回给客户端
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	//这里js是[]byte类型,可以直接append rune类型
	js = append(js, '\n')

	//可以稍后自定义一个响应头headers，然后直接将内容遍历写到响应头
	for key, value := range headers {
		w.Header()[key] = value
	}
	//内容可以自定义，默认是Content-Type: text/plain; charset=utf-8，返回json数据时一般设置为application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}
