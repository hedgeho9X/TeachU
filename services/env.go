package services

const (
	DoubaoLite    = "ep-20250311120726-h7xml"
	AnalyzePrompt = "识别试卷内容，提取每道题目的题号、知识点和题目内容。返回格式要求：每道题目占一行，格式为：{'question_id':'', 'key':'','content':''。返回 JSON 格式。当题目为空时返回空json"
)
