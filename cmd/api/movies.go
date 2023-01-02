package main

import (
	"errors"
	"fmt"
	"greenlight/internal/validator"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// 剥去字段值的双引号
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// 分离值得到一个string数组，期望的结果应该是 xxx mins 这两个元素
	parts := strings.Split(unquotedJSONValue, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// 将 xxx 转为int32类型以与Runtime的潜在类型一致
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// 将值传给r，这也就意味着更改了r的值
	*r = Runtime(i)
	return nil
}
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	//声明一个匿名结构体，用来储存decoder的内容，
	//结构体的数据项须可导出，因为在decode过程中就意味着json包在使用数据项
	//数据项的命名需要和希望decode的内容一一对应，否则会被忽略
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime Runtime  `json:"runtime"`
		Genres  []string `json:"genres"`
	}
	//创建一个json.Decoder对象从请求体中读取内容，
	//然后使用decode方法将读取的内容decode到&input(非空指针)
	//r.Body在生成decoder后会被http.server自动close
	err := app.readJSON(w, r, &input)
	//err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}
	v := validator.New()
	ValidateMovie(v, movie)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	//when printing structs, the plus flag (%+v) adds field names
	fmt.Fprintf(w, "%+v\n", input)
}
func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		//	解析不了，那么可能是请求的id并非数字,返回一个404
		//http.NotFound(w, r)
		app.notFoundResponse(w, r)
		//http.Error(w, err.Error(), http.StatusBadRequest)
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
		app.serverErrorResponse(w, r, err)
	}
}
