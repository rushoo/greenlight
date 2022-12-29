package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) route() *httprouter.Router {
	router := httprouter.New()

	//httprouter通过这种方式，处理意外的请求，比如请求路径不存在时，就可以通过这种方式返回
	//自定义的输出结果app.notFoundResponse
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	return router
}
