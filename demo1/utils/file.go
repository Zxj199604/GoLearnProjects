package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// AppendToFile 在文件末尾追加内容
func AppendToFile(filename string, content string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("无法打开文件：%s", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("无法写入文件：%s", err)
	}
	return nil
}

// WriteFile 覆盖写文件
func WriteFile(filename string, content string) error {
	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("无法写入文件：%s", err)
	}
	return nil
}

// RemoveFileOrDir 删除文件或文件夹
func RemoveFileOrDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("无法删除文件或文件夹：%s", err)
	}
	return nil
}

// FieldExistsInFile 检查文件中是否至少存在一个字段
func FieldExistsInFile(filename string, fields []string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, fmt.Errorf("无法打开文件：%s", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, field := range fields {
			if strings.Contains(line, field) {
				return true, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("读取文件时发生错误：%s", err)
	}

	return false, nil
}
