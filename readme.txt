EduSpark 项目概览
【请注意：主要子文件夹下同样包含一个readme.txt解释该文件夹作用，您可以在感兴趣的文件夹下继续阅览】
主要文件夹结构说明
本项目采用模块化的结构组织代码，主要文件夹及其作用如下：

1.  cmd:
    作用: 包含项目的主程序入口 (main.go)。负责初始化配置、数据库连接、Redis 连接以及启动 Web 服务器。

2.  internal:
    作用: 存放项目的核心业务逻辑代码，遵循 Go 的 internal 目录规范，确保内部代码的封装性。
    - config: 存放项目配置信息，如数据库连接、JWT 密钥、Redis 地址、外部服务 API Key 等。
    - controllers: 处理 HTTP 请求，作为 MVC 架构中的控制层。负责解析请求、调用 services 处理逻辑并返回响应。
    - middlewares: 存放 Gin 中间件，用于处理请求过程中的通用任务，如 JWT 身份验证、日志记录等。
    - models: 定义数据模型（Go 结构体），对应数据库表结构或 API 的数据格式。
    - services: 包含应用的核心业务逻辑。被 controllers 调用，负责执行具体业务规则、数据处理、与数据库交互以及调用外部服务等。

3.  routes:
    作用: 定义项目的所有 API 路由。使用 Gin 框架将 HTTP 请求映射到相应的 controllers 处理函数。包含路由分组、中间件应用等。

4.databases
    作用：数据库的创建sql文件