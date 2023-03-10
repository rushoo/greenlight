package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	//一种方式是原生的golang数据类型(包括map、slice、struct)序列化为json
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}
	err := app.writeJSON(w, http.StatusOK, envelope{"info:": data}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
