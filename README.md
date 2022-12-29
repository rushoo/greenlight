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

### 第二步
1、更新程序使客户端得以返回json格式响应
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
2、请求`localhost:4000/v1/movies/:id`返回json格式movies信息
