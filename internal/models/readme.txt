EduSpark Models Directory 说明

文件夹作用:

models 文件夹定义了 EduSpark 项目中使用的数据结构。这些结构通常代表数据库中的表、API 请求/响应体或应用程序内部使用的核心数据实体。它们为 service 层和 data access 层提供了统一的数据格式。

各主要模型文件描述:

1.  user.go:
    作用: 定义 `User` 结构体，包含用户账户的基本信息，如 ID、手机号、邮箱、用户名、密码哈希等。

2.  class.go:
    作用: 定义 `Class` 结构体，表示一个班级，包含班级编号、年级、创建者 ID 等信息，并关联学生列表。

3.  student.go:
    作用: 定义 `Student` 结构体，表示一个学生，包含学生姓名、学号以及所属班级 ID。

4.  exam.go:
    作用: 定义 `Exam` 结构体，表示一次考试事件，包含考试名称、科目、创建用户 ID、班级 ID 等。

5.  problems.go:
    作用: 定义 `Problems` 结构体，表示具体的试题，包含题目内容、所属考试 ID、知识点、题号、总分等。

6.  score.go:
    作用: 定义 `Score` 结构体及相关辅助结构（如 `QuestionScore`, `StudentSimple`, `StudentScoreResponse`），用于记录学生在特定考试或题目上的得分情况。

7.  analysis.go:
    作用: 定义与学情分析相关的结构体（如 `StudentAnalysisResponse`, `StudentKeypoints`, `StudentHistory`），用于封装学生的考试表现、知识点掌握情况和历史成绩等分析结果。

8.  resource.go:
    作用: 定义 `Resource` 结构体，表示教学资源，包含文件名、对象存储键（ObjectKey）、文件大小、类型、所属学科和年级等信息。

9.  jwt.go:
    作用: 定义 `Claims` 结构体，用于构建 JWT (JSON Web Tokens) 的载荷，包含用户 ID、用户名等身份验证信息。

10. email.go:
    作用: 定义 `EmailConfig` 结构体，用于存储发送邮件所需的配置信息，如发件人地址、密码、SMTP 服务器地址和端口等。