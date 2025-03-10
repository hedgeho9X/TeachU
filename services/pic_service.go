package services

import (
	"github.com/volcengine/volc-sdk-golang/service/visual"
)

// PicGenerateRequest 图片生成请求参数
type PicGenerateRequest struct {
	Prompt string `json:"prompt"` // 提示词
	Width  int    `json:"width"`  // 图片宽度
	Height int    `json:"height"` // 图片高度
	Seed   int    `json:"seed"`   // 随机种子
}

func PicGenerate(req PicGenerateRequest) (map[string]interface{}, error) {
	testAk := "AKLTY2RhZjI5OTQ0YzhhNDQ1MDgxODVlMWJjNTVmYTJhNTE"
	testSk := "TjJReU9EZ3laalF3WW1ZeU5HUTJPV0l5WmpWaFpHVTFZMlpoWkdObE1XRQ=="

	visual.DefaultInstance.Client.SetAccessKey(testAk)
	visual.DefaultInstance.Client.SetSecretKey(testSk)
	//visual.DefaultInstance.SetRegion("region")
	//visual.DefaultInstance.SetHost("host")

	//请求Body(查看接口文档请求参数-请求示例，将请求参数内容复制到此)
	reqBody := map[string]interface{}{
		"req_key":           "high_aes_general_v21_L",
		"prompt":            req.Prompt,
		"model_version":     "general_v2.1_L",
		"req_schedule_conf": "general_v20_9B_pe",
		"llm_seed":          -1,
		"seed":              req.Seed,
		"scale":             3.5,
		"ddim_steps":        25,
		"width":             req.Width,
		"height":            req.Height,
		"use_pre_llm":       true,
		"use_sr":            true,
		"sr_seed":           -1,
		"sr_strength":       0.4,
		"sr_scale":          3.5,
		"sr_steps":          20,
		"is_only_sr":        false,
		"return_url":        true,
		// "logo_info": map[string]interface{}{
		// 	"add_logo":          false,
		// 	"position":          0,
		// 	"logo_text_content": "这里是明水印内容",
		// 	"language":          0,
		// 	"opacity":           0.3,
		// },
	}

	pics, _, err := visual.DefaultInstance.CVProcess(reqBody)
	if err != nil {
		return nil, err
	}
	return pics, nil
}
