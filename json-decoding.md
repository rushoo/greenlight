# 处理json请求

### json decoding
类似于json encoding, 有两种方式可以decode json into go object, `json.Decoder`和`json.Unmarshal()`   
一般而言，decoding JSON from a HTTP request body使用`json.Decoder`会更方便一些，代码精简且效率更高。  
```
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
    // 申明一个匿名结构体，将用来存储浏览器请求体中的json数据
    var input struct {
        Title string `json:"title"`
        Year int32 `json:"year"`
        Runtime int32 `json:"runtime"`
        Genres []string `json:"genres"`
    }
    
    // 初始化一个json decoder,然后将请求体的内容decode到input结构体的指针
    // When decoding a JSON object into a struct, JSON 中字段的内容会依据标签map到结构体每个字段里
    // 不能正确mapping的字段会被略过，返回值是error类型，这里省略了错误处理代码段
    _ := json.NewDecoder(r.Body).Decode(&input)  
    
    // 将请求体的内容写回客户端, %+v 用以结果中显示结构体字段名称
    fmt.Fprintf(w, "%+v\n", input)
}
```   
查看json decode的结果，第二次请求中year没给值，所以返回的结果中那一项是`Year:0`      
```
# -d发送POST请求，-i参数返回响应头信息
$ BODY='{"title":"Moana","year":2016,"runtime":107, "genres":["animation","adventure"]}'
$ curl -i -d "$BODY" localhost:4000/v1/movies            
HTTP/1.1 200 OK                                                 
Date: Mon, 02 Jan 2023 07:15:46 GMT                             
Content-Length: 65                                              
Content-Type: text/plain; charset=utf-8                         
                                                                
{Title:Moana Year:2016 Runtime:107 Genres:[animation adventure]}


$ BODY2='{"title":"Moana","runtime":107, "genres":["animation","adventure"]}'
$ curl -d "$BODY2" localhost:4000/v1/movies                                       
{Title:Moana Year:0 Runtime:107 Genres:[animation adventure]}
```
### Managing Bad Requests
现在存在这样几个问题需要改进：   
+ 请求体非json格式，比如xml格式的
+ 请求体的json格式含错误
+ 请求体中字段类型与decode目标容器的字段类型不匹配
+ 请求体为空   

当这些情况发生时,程序会直接报错，但错误信息不够明了，需要做进一步处理。
```
# xml请求
$ curl -d '<?xml version="1.0" encoding="UTF-8"?><note><to>Alice</to></note>' localhost:4000/v1/movies
{
"error": "invalid character '\u003c' looking for beginning of value"
}

# 格式错误
$ curl -d '{"title": "Moana", }' localhost:4000/v1/movies
{
"error": "invalid character '}' looking for beginning of object key string"
}

# 格式错误
$ curl -d '["foo", "bar"]' localhost:4000/v1/movies
{
"error": "json: cannot unmarshal array into Go value of type struct { Title string
\"json:\\\"title\\\"\"; Year int32 \"json:\\\"year\\\"\"; Runtime int32 \"json:\\
\"runtime\\\"\"; Genres []string \"json:\\\"genres\\\"\" }"
}

# 字段类型不匹配
$ curl -d '{"title": 123}' localhost:4000/v1/movies
{
"error": "json: cannot unmarshal number into Go struct field .title of type string"
}

# 空请求
$ curl -X POST localhost:4000/v1/movies
{
"error": "EOF"
}
```

`json.NewDecoder(r.Body).Decode(&input)`返回以下数种错误类型:   

| 错误类型                       | 原因               |
|----------------------------|------------------|
| io.EOF                     | 读取错误，请求体内容为空     |
| io.ErrUnexpectedEOF        | 读取错误，一般是符号不对称    |
| json.SyntaxError           | json处理错误，内容含语法错误 |
| json.UnmarshalTypeError    | JSON处理错误，字段类型不匹配 |
| json.InvalidUnmarshalError | JSON处理错误，函数传值有问题 |

```
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	err := json.NewDecoder(r.Body).Decode(dst)

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
	   所以这里对前两种错误使用的是errors.Is值比较，后三种使用errors.As类型比较
	*/
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

        // 根据不同的错误类型完善错误输出结果
		switch {
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.As(err, &invalidUnmarshalError):
            //函数传参错误属开发错误，不应该出现的，panic提醒
			panic(err)
		default:
			return err
		}
	}
	return nil
}
```
此时错误信息显示变得更加详细了：   
```
# xml格式请求
$ curl -d '<?xml version="1.0" encoding="UTF-8"?><note><to>Alex</to></note>' localhost:4000/v1/movies
{
"error": "body contains badly-formed JSON (at character 1)"
}

# 错误json格式（逗号）
$ curl -d '{"title": "Moana", }' localhost:4000/v1/movies
{
"error": "body contains badly-formed JSON (at character 20)"
}

# 错误json格式
$ curl -d '["foo", "bar"]' localhost:4000/v1/movies
{
"error": "body contains incorrect JSON type (at character 1)"
}

# 字段类型不匹配 'title'
$ curl -d '{"title": 123}' localhost:4000/v1/movies
{
"error": "body contains incorrect JSON type for \"title\""
}

# 空请求体
$ curl -X POST localhost:4000/v1/movies
{
"error": "body must not be empty"
}
```

### Restricting Inputs
前面对于请求体的格式做了处理，完善了无效格式的错误提示，但仅限于格式合法而并未校验输入的具体内容，比如在请求体中添加   
无关内容或者缺乏重要字段的内容，还是正常地响应：   
```
$ curl -i -d '{"title": "Moana", "rating":"PG"}' localhost:4000/v1/movies
HTTP/1.1 200 OK
Date: Tue, 06 Apr 2021 18:51:50 GMT
Content-Length: 41
Content-Type: text/plain; charset=utf-8
{Title:Moana Year:0 Runtime:0 Genres:[]}
```
对于请求体中的未知字段可以使用`dec.DisallowUnknownFields()`方法加以限制，同时也可以通过`http.MaxBytesReader`    
限制请求内容小于1M。json.Decode()每次用readValue()方法仅读取一条json数据，可以在decode方法后再调用一次，如果    
希望请求体仅一条json，那么就会产生io.EOF错误。具体实现如下(这里仅显示新增代码)：   
```
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


// 针对http.MaxBytesReader增加错误处理逻辑
case err.Error() == "http: request body too large":
    return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
//	针对dec.DisallowUnknownFields()添加一条处理逻辑
case strings.HasPrefix(err.Error(), "json: unknown field "):
	fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
	return fmt.Errorf("body contains unknown key %s", fieldName)

//要限制请求体一次仅含一条数据，当第一条json decode结束后继续decode，期望io.EOF
err = dec.Decode(&struct{}{})
if !errors.Is(err, io.EOF) {
	return errors.New("body must only contain a single JSON value")
}
```
新的处理结果符合，但关键字段内容校验还需要完善：   
```
# 内容过大，提前准备好的largefile.json文件
$ curl -d largefile.json localhost:4000/v1/movies
{
"error": "body must not be larger than 1048576 bytes"
}

# 未知字段
$ curl -d '{"title": "Moana", "rating":"PG"}' localhost:4000/v1/movies
{
        "error": "body contains unknown key \"rating\""
}

# 单条json后还有内容
$ curl -d '{"title": "Moana"}{"title": "Top Gun"}' localhost:4000/v1/movies
{
        "error": "body must only contain a single JSON value"
}
$ curl -d '{"title": "Moana"} :~()' localhost:4000/v1/movies
{
        "error": "body must only contain a single JSON value"        
}

# 关键字段内容为空，在当前是合法的
$ curl -d '{"title": ""} ' localhost:4000/v1/movies
{Title: Year:0 Runtime:0 Genres:[]}
```

### Validating JSON Input    
接下来对输入的内容做些检查，要求：   
+ 标题不为空且字长小于200
+ 年代不为空且在1888-now之间
+ 电影时长不为空且为正整数
+ 题材描述不为空、不重复、小于5
当发生以上校验错误是，返回一个422状态码码和相关错误信息，意思是请求格式正确，但是由于含有语义错误，无法响应。这些行为   
可以在`app.readJSON(w, r, &input)`之后，分别对结构体的每个字段获得的具体内容加以校验而实现。关键代码如下：   
```
type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}


	...
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

//=============================================================

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
通过以上这段代码逻辑可以实现对于输入数据的校验，只不过这种校验是发生在decode之后的，对于目标容器获得请求体的内容之后，    
对结构体某个字段的具体值的校验。   

### Custom JSON Decoding
假设对于这样一个请求：
```
$ curl -d '{"title": "Moana", "runtime": "107 mins"}' localhost:4000/v1/movies
{
"error": "body contains incorrect JSON type for \"runtime\""
}
```    
类似于json.Encoder, 在decode时也可以通过json.Unmarshaler自定义unmarshal行为。前面增加字段内容校验时，在复制    
input的值到movie对象时已经将Runtime的类型改成了Runtime，从而与Movie的字段类型一致。下面对于Runtime类型重写    
`UnmarshalJSON(jsonValue []byte) error`方法以实现自定义行为：    
```
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
```
结果符合预期：   
```
$ BODY='{"title":"","year":1000,"runtime":"-123 mins","genres":["sci-fi","sci-fi"]}'
$ curl -i -d "$BODY" localhost:4000/v1/movies
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json
Date: Mon, 02 Jan 2023 16:38:50 GMT
Content-Length: 180
{
    "error": {
        "genres": "must not contain duplicate values",
        "runtime": "must be a positive integer",
        "title": "must be provided",
        "year": "must be greater than 1888"
    }
}

$  BODY='{"title":"Moana","year":2016,"runtime":"107 mins","genres":["animation","adventure"]}' && 
curl -i -d "$BODY" localhost:4000/v1/movies
HTTP/1.1 200 OK                                                 
Date: Mon, 02 Jan 2023 16:40:25 GMT                             
Content-Length: 65                                              
Content-Type: text/plain; charset=utf-8                         

{Title:Moana Year:2016 Runtime:107 Genres:[animation adventure]}
```
