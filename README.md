# 简介
接下来会创建一个json api，名字叫Greenlight，支持查询和管理有关电影的信息。大概有以下功能：
<table>
  <tr>
    <th>Method</th>
    <th>URL Pattern</th>
    <th>Action</th>
  </tr>
  <tr>
    <td>GET</td>
    <td>/v1/healthcheck</td>
    <td>Show application health and version information</td>
  </tr>
  <tr>
    <td>GET</td>
    <td>/v1/movies</td>
    <td>Show the details of all movies</td>
  </tr>
  <tr>
    <td>POST</td>
    <td>/v1/movies</td>
    <td>Create a new movie</td>
  </tr>
  <tr>
    <td>GET</td>
    <td>/v1/movies/:id</td>
    <td>Show the details of a specific movie</td>
  </tr>
  <tr>
    <td>PATCH</td>
    <td>/v1/movies/:id</td>
    <td>Update the details of a specific movie</td>
  </tr>
  <tr>
    <td>DELETE</td>
    <td>/v1/movies/:id</td>
    <td>Delete a specific movie</td>
  </tr>
  <tr>
    <td>POST</td>
    <td>/v1/users</td>
    <td>Register a new user</td>
  </tr>
  <tr>
    <td>PUT</td>
    <td>/v1/users/activated</td>
    <td>Activate a specific user</td>
  </tr>
  <tr>
    <td>PUT</td>
    <td>/v1/users/password</td>
    <td>Update the password for a specific user</td>
  </tr>
  <tr>
    <td>POST</td>
    <td>/v1/tokens/authentication</td>
    <td>Generate a new authentication token</td>
  </tr>
  <tr>
    <td>POST</td>
    <td>/v1/tokens/password-reset</td>
    <td>Generate a new password-reset token</td>
  </tr>
  <tr>
    <td>GET</td>
    <td>/debug/vars</td>
    <td>Display application metrics</td>
  </tr>
</table>

### 第一步
- 1、 搭建项目框架
```
.
├── bin             //将用来存放编译好的二进制文件，用以部署服务
├── cmd            
│ └── api           //该目录下存放核心代码文件
│ └── main.go
├── internal        //仅限本项目公用的代码，外部不可引用该包
├── migrations      //SQL版本管理
├── remote          //生产环境的的配置文件和相关脚本
├── go.mod          //项目依赖管理
└── Makefile        //一些build或者其它工具脚本
```
- 2、启动一个简单的http服务以监听请求
- 3、通过命令行管理配置设置以及使用依赖注入
```
var cfg config
flag.IntVar(&cfg.port, "port", 4000, "API  port")
flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
flag.Parse()
//省略了相关结构体定义
app := &application{
	config: cfg,
	logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
}
//自定义server以使用自定义的port
srv := &http.Server{
	Addr:    fmt.Sprintf(":%d", cfg.port),
	Handler: app.route(),
}
log.Fatal(srv.ListenAndServe())
```
- 4、引入`httprouter`包以实现restful接口


两个问题：
1、并未真正使用自定义的日志

### 第二步，响应json数据
- 1、更新程序使客户端得以返回json格式响应
```
//一种方式是直接以json格式定义好数据
js := `{"status": "available", "environment": %q, "version": %q}`
js = fmt.Sprintf(js, app.config.env, version)
w.Write([]byte(js))

//一种方式是原生的golang数据类型(包括map、slice、struct)序列化为json
data := map[string]string{
	"status":      "available",
	"environment": app.config.env,
	"version":     version,
}
js, _ := json.Marshal(data)  //省略了错误处理和响应头设置
w.Write(js)
```
- 2、请求`localhost:4000/v1/movies/:id`返回json格式movies信息    
- 3、丰富错误输出信息：   
定义方法返回json格式自定义错误；        
通过httprouter的`router.NotFound = http.HandlerFunc(app.notFoundResponse)`这种方式对于非预期请求使用自定义错误处理

### 第三步，处理json请求
1、接受接送请求，将请求体的内同decode到变量，但仅仅这样对于处理不了的或者是错误的请求，响应内容还不够明晰。   
```
//声明一个匿名结构体，用来储存decoder的内容，
//结构体的数据项须可导出，因为在decode过程中就意味着json包在使用数据项
//数据项的命名需要和希望decode的内容一一对应，否则会被忽略
var input struct {
	Title   string   `json:"title"`
	Year    int32    `json:"year"`
	Runtime int32    `json:"runtime"`
	Genres  []string `json:"genres"`
}
//创建一个json.Decoder对象从请求体中读取内容，
//然后使用decode方法将读取的内容decode到&input(非空指针)
//r.Body在生成decoder后会被http.server自动close
_ := json.NewDecoder(r.Body).Decode(&input) //隐藏了错误处理
fmt.Fprintf(w, "%+v\n", input)
```
2、增加请求内容长度和未知字段限制，根据decode过程中可能的错误类型，以switch-case方式完善错误日志。
```
maxBytes := 1_048_576
r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
err := dec.Decode(dst)

// 限制请求体内容为仅一条json
err = dec.Decode(&struct{}{})
if !errors.Is(err, io.EOF) {
    return errors.New("body must only contain a single JSON value")
}
```
3、增加对于格式合法的输入内容的校验，比如关键字非空以及数据范围，通过对于decode后具体字段结果的检查实现。
```
// 将容器input每个字段的值相应地复制到movie中，再检查movie是否符合要求
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
```
4、自定义unmarshal的行为，通过对字段类型重写`UnmarshalJSON(jsonValue []byte) error`方法来实现。    
```
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
```



