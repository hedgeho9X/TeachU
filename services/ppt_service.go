package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// 鉴权签名生成
func GenerateSignature(auth *AuthConfig, timestamp int64) string {
	// 生成MD5
	md5Str := fmt.Sprintf("%s%d", auth.AppID, timestamp)
	hasher := md5.New()
	hasher.Write([]byte(md5Str))
	authMD5 := hex.EncodeToString(hasher.Sum(nil))

	// 生成HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(auth.Secret))
	mac.Write([]byte(authMD5))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// 创建带鉴权的请求
func createRequest(auth *AuthConfig, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	signature := GenerateSignature(auth, timestamp)

	req.Header.Add("appId", auth.AppID)
	req.Header.Add("timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Add("signature", signature)

	return req, nil
}

// 查询PPT主题列表
func GetPPTThemeList(auth *AuthConfig, reqData PPTThemeRequest) (*CommonResponse, error) {
	url := "https://zwapi.xfyun.cn/api/ppt/v2/template/list"

	jsonData, _ := json.Marshal(reqData)
	request, _ := createRequest(auth, "POST", url, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result CommonResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}

// 生成PPT（直接生成）
func GeneratePPT(auth *AuthConfig, reqData GeneratePPTRequest) (*CommonResponse, error) {
	url := "https://zwapi.xfyun.cn/api/ppt/v2/create"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 处理不同输入类型
	if reqData.File != nil {
		part, _ := writer.CreateFormFile("file", reqData.FileName)
		io.Copy(part, reqData.File)
	} else if reqData.FileURL != "" {
		writer.WriteField("fileUrl", reqData.FileURL)
	} else {
		writer.WriteField("query", reqData.Query)
	}

	// 添加其他字段
	if reqData.FileName != "" {
		writer.WriteField("fileName", reqData.FileName)
	}
	if reqData.TemplateID != "" {
		writer.WriteField("templateId", reqData.TemplateID)
	}
	if reqData.Author != "" {
		writer.WriteField("author", reqData.Author)
	}

	writer.WriteField("isCardNote", fmt.Sprintf("%t", reqData.IsCardNote))
	writer.WriteField("search", fmt.Sprintf("%t", reqData.Search))
	writer.WriteField("isFigure", fmt.Sprintf("%t", reqData.IsFigure))
	writer.WriteField("aiImage", reqData.AIImage)
	writer.WriteField("language", reqData.Language)

	writer.Close()

	request, _ := createRequest(auth, "POST", url, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CommonResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}

// 查询PPT生成进度
func GetPPTProgress(auth *AuthConfig, sid string) (*CommonResponse, error) {
	url := fmt.Sprintf("https://zwapi.xfyun.cn/api/ppt/v2/progress?sid=%s", sid)

	request, _ := createRequest(auth, "GET", url, nil)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CommonResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}

// func main() {
// 	// 当前硬编码配置
// 	- auth := &services.AuthConfig{
// 	-    AppID:  "8a1fff11",
// 	-    Secret: "NDFkYzU1MDZmODY0Y2ZhNTgzYTg1OTU0",
// 	- }
//
// 	// 建议改为配置加载
// 	+ var config struct {
// 	+    PPT struct {
// 	+        AppID  string
// 	+        Secret string
// 	+    } `mapstructure:"ppt"`
// 	+ }
// 	+ viper.Unmarshal(&config)
//
// 	// 示例1：查询PPT主题列表
// 	themeReq := PPTThemeRequest{
// 		Style:    "商务",
// 		PageSize: 10,
// 	}
// 	themeRes, _ := GetPPTThemeList(auth, themeReq)
// 	fmt.Printf("主题列表响应: %+v\n", themeRes)

// 	// 示例2：直接生成PPT
// 	pptReq := GeneratePPTRequest{
// 		Query:      "计算机网络",
// 		Language:   "cn",
// 		IsFigure:   true,
// 		AIImage:    "normal",
// 		IsCardNote: true,
// 	}
// 	pptRes, _ := GeneratePPT(auth, pptReq)
// 	fmt.Printf("PPT生成响应: %+v\n", pptRes)

// 	// 示例3：查询生成进度
// 	if pptRes.Data != nil {
// 		data := pptRes.Data.(map[string]interface{})
// 		sid := data["sid"].(string)

// 		// 轮询检查PPT生成进度，每3秒检查一次，最多检查20次
// 		for i := 0; i < 1000; i++ {
// 			progressRes, err := GetPPTProgress(auth, sid)
// 			if err != nil {
// 				fmt.Printf("查询进度出错: %v\n", err)
// 				break
// 			}

// 			fmt.Printf("生成进度: %+v\n", progressRes)

// 			// 检查是否生成完成
// 			if progressRes.Data != nil {
// 				pptStatus := progressRes.Data.(map[string]interface{})["pptStatus"].(string)
// 				if pptStatus == "done" { // 假设状态2表示生成完成
// 					fmt.Println("PPT生成完成！")
// 					break
// 				}
// 			}
// 		}

// 		time.Sleep(3 * time.Second)
// 	}
// }
