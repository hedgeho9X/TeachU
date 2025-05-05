EduSpark Controllers Directory 说明

文件夹作用:

controllers 文件夹是 EduSpark 项目的入口点，负责处理传入的 HTTP 请求。它充当了 MVC (Model-View-Controller) 架构中的控制层角色。Controllers 解析请求参数，调用相应的 services 层来执行业务逻辑，然后格式化并返回 HTTP 响应给客户端。它们通常还负责处理用户身份验证、输入验证和错误处理等任务。

各主要 Controller 文件描述:

1.  user_controller.go: 
    作用: 处理用户认证相关的 HTTP 请求，例如用户注册、登录、获取用户信息、请求密码重置、验证邮箱/手机验证码等。调用 `user_service`。

2.  class_controller.go:
    作用: 处理班级和学生管理的 HTTP 请求，包括创建班级、获取用户拥有的班级列表、删除班级、添加/导入学生、列出班级内学生、删除学生等。调用 `class_service` 和 `student_service`。

3.  exam_controller.go:
    作用: 处理考试相关的 HTTP 请求，如创建新考试（接收上传的试卷文件）、解析试卷内容（调用 AI 服务）、获取指定班级的考试列表、删除考试等。调用 `exam_service`。

4.  problems_controller.go:
    作用: 处理试卷题目相关的 HTTP 请求，主要是创建（或批量添加）与特定考试关联的试题信息（知识点、题号、分数、内容）。调用 `problems_services`。

5.  score_controller.go:
    作用: 处理学生成绩相关的 HTTP 请求，包括上传（批量录入）某次考试的学生得分、获取某次考试的成绩列表等。调用 `score_service`。

6.  analysis_controller.go:
    作用: 处理学情分析数据的 HTTP 请求，提供基于已有数据的统计分析结果，例如获取某次考试的班级整体分析、学生个人分析、学生排名等。调用 `analysis_service`。

7.  ai_analysis_controller.go:
    作用: 处理由 AI 生成的学情分析报告的 HTTP 请求，例如请求 AI 对班级或单个学生的考试表现进行深入分析和总结。调用 `analysis_service` 中的 AI 分析功能。

8.  ai_controller.go:
    作用: 处理通用的 AI 功能相关的 HTTP 请求，例如与 AI 助教进行聊天交互、请求 AI 生成教学 PPT 等。调用 `chat_service`、`ppt_service` 等。

9.  rag_controller.go:
    作用: 处理基于 RAG (Retrieval-Augmented Generation) 技术的智能推荐请求，如根据学生学情分析结果或用户输入的消息，生成个性化的学习建议或资源推荐。调用 `rag_service`。

10. resource_controller.go:
    作用: 处理教学资源管理的 HTTP 请求，可能包括上传资源文件、搜索/查询资源、获取资源访问链接等。调用 `resource_service` 和 `oss_service`。
