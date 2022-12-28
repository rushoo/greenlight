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
- 2、启动一个简单的http服务以监听请求
- 3、通过命令行管理配置设置以及使用依赖注入
- 4、引入`httprouter`包以实现restful接口



