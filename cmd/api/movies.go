package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

//type Movie struct {
//	ID        int64     // ID
//	CreatedAt time.Time // 电影记录创建时间
//	Title     string    // 标题
//	Year      int32     // 上映年代
//	Runtime   int32     // 电影播放时长
//	Genres    []string  // 电影主题，包括(romance, comedy, etc.)
//	Version   int32     // 从1开始，每次更新时+1
//}

//type Movie struct {
//	ID        int64     `json:"id"`
//	CreatedAt time.Time `json:"created_at"`
//	Title     string    `json:"title"`
//	Year      int32     `json:"year"`
//	Runtime   int32     `json:"runtime"`
//	Genres    []string  `json:"genres"`
//	Version   int32     `json:"version"`
//}

//type Movie struct {
//	ID        int64     `json:"id"`
//	CreatedAt time.Time `json:"-"` // - 不显示
//	Title     string    `json:"title"`
//	Year      int32     `json:"year,omitempty"`           // 增加 omitempty 空不显示
//	Runtime   int32     `json:"runtime,omitempty,string"` // 增加 string 指定以string类型显示
//	Genres    []string  `json:"genres,omitempty"`
//	Version   int32     `json:"version"`
//}

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // - 不显示
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`           // 增加 omitempty 空不显示
	Runtime   Runtime   `json:"runtime,omitempty,string"` // 增加 string 指定以string类型显示
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}
type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	//需要使用双引号括起来，否则就是一个无效的json字符串导致序列化失败
	quotedJsonValue := strconv.Quote(jsonValue)
	return []byte(quotedJsonValue), nil
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movies")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		//	解析不了，那么可能是请求的id并非数字,返回一个404
		//http.NotFound(w, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//使用一个movie结构体用来传递movie信息
	movie := Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}
	err = app.writeJSON(w, 200, envelope{"movie": movie}, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
