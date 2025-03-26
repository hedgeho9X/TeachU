package main

import (
	"flag"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/Hedgeho9X/TeachU/services"
)

// 创建模拟文件头结构（添加错误处理）
func createMockFileHeader(filePath string) (*multipart.FileHeader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("文件打开失败: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	return &multipart.FileHeader{
		Filename: filepath.Base(filePath),
		Size:     stat.Size(),
		Header:   make(map[string][]string),
	}, nil
}

// main函数中需要添加错误处理
func main() {
	var testFilePath string
	flag.StringVar(&testFilePath, "file", "", "要解析的Word文件路径")
	flag.Parse()

	if testFilePath == "" {
		fmt.Println("必须使用 -file 参数指定测试文件")
		return
	}

    // 创建模拟的multipart.FileHeader
    fileHeader, err := createMockFileHeader(testFilePath)
    if err != nil {
        fmt.Printf("文件准备失败: %v\n", err)
        return
    }

	// 调用解析函数
	content, err := services.ProcessUploadedFile(fileHeader)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}

	fmt.Println("========= 解析结果 =========")
	fmt.Printf("内容长度: %d 字符\n", len(content))
	fmt.Println("========= 内容预览 =========")
	if len(content) > 500 {
		fmt.Println(content[:500] + "...")
	} else {
		fmt.Println(content)
	}
}
