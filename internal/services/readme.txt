EduSpark Services Directory 说明

文件夹作用:

services 文件夹是 EduSpark 项目核心业务逻辑的所在地。它包含了处理应用程序主要功能的代码，通常被 controllers 层调用。这里的服务负责执行具体的业务规则、数据处理、与数据库交互以及调用外部服务等。

各主要服务文件描述:

1.  user_service.go:
    作用: 处理用户相关的业务逻辑，包括用户注册、登录验证、用户信息获取、密码重置、邮箱/手机验证码服务等。

2.  class_service.go:
    作用: 管理班级相关的操作，如创建班级、获取用户班级列表、删除班级等。

3.  student_service.go:
    作用: 管理班级内的学生信息，包括添加学生、列出学生、删除学生等。

4.  exam_service.go:
    作用: 负责考试相关的功能，包括创建考试、通过 AI (文本/图片) 解析试卷题目、获取考试列表、删除考试、解析试题 JSON 等。

5.  analysis_service.go:
    作用: 提供学情分析相关的服务，例如分析学生的考试表现、知识点掌握情况等。

6.  resource_service.go:
    作用: 管理教学资源，提供资源的搜索、查询以及通过 OSS 获取访问链接等功能。

7.  ppt_service.go:
    作用: 对接外部 API (讯飞)，实现 PPT 的自动生成、进度查询等功能。

8.  jwt_service.go:
    作用: 负责生成和管理 JWT (JSON Web Tokens)，用于用户身份验证和授权。

9.  rag_service.go:
    作用: 实现 RAG功能，用于基于学情分析的智能推荐。

10. oss_services.go:
    作用: 封装与阿里云 OSS (对象存储服务) 的交互逻辑，用于文件上传、下载、获取访问 URL 等。

11. chat_service.go:
    作用: 处理应用内的聊天或问答相关功能。

12. score_service.go:
    作用: 管理和处理分数、评分相关的业务逻辑。

13. problems_services.go:
    作用: 处理与具体题目（可能是练习题、考试题）相关的业务逻辑。

14. pic_service.go:
    作用: 提供图片处理相关的服务。
