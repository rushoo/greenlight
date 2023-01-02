package main

import (
	"fmt"
	"net/http"
)

// 服务端记下关于请求的错误信息
func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

// 向客户端发送json格式的自定义错误，以及对应的状态码
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	envErr := envelope{"error": message}
	err := app.writeJSON(w, status, envErr, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	//服务端错误，需要记录下来
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
