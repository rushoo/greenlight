package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
	"strings"
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

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	//js, err := json.Marshal(data)
	js, err := json.MarshalIndent(data, "", "\t")
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
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	/*
	   Go 1.13 errors新增特性含三个新函数：Is、As、Unwrap
	   简单来说，errors.Is判断两个值是否相等,errors.As判断两个类型是否一致。
	   Unwrap方法将嵌套的 error 解析出来，多层嵌套需要调用多次才能获取最里层的error。
	   两者判断过程中都可能会调用errors.Unwrap方法去取底层的err值。

	   就以上decode错误处理过程而言，涉及到的错误分别来源于以下定义：
	   var EOF = errors.New("EOF")
	   var ErrUnexpectedEOF = errors.New("unexpected EOF")
	   type SyntaxError struct {
	   	msg    string
	   	Offset int64
	   }
	   type UnmarshalTypeError struct {
	   	Value  string
	   	Type   reflect.Type
	   	Offset int64
	   	Struct string
	   	Field  string
	   }
	   type InvalidUnmarshalError struct {
	   	Type reflect.Type
	   }
	*/
	// 使用http.MaxBytesReader()方法限制请求体内容最大为 1MB.
	// 当超限时http包返回错误http.Error(), return "http: request body too large"
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// 使用DisallowUnknownFields()对于限制请求体中使用未知字段，
	// 可能产生这条错误逻辑 fmt.Errorf("json: unknown field %q", key)
	// dec.Decode通过dec.readValue()方法读取第一条json数据
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	//err := json.NewDecoder(r.Body).Decode(dst)

	// 根据不同的错误类型完善错误输出结果
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// 针对http.MaxBytesReader增加错误处理逻辑
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &syntaxError):
			// 显示JSON请求中语法错误的位置
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		//	针对dec.DisallowUnknownFields()添加一条处理逻辑
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &invalidUnmarshalError):
			//函数传参错误属开发错误，不应该出现的，panic提醒
			panic(err)
		default:
			return err
		}
	}
	//要限制请求体一次仅含一条数据，当第一条json decode结束后继续decode，期望io.EOF
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}
