package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

func Test_exec(t *testing.T) {
	result := ExecuteCommand("/Users/zxj", "ps aux")
	if result.Err != nil {
		fmt.Println("命令执行失败:", result.Err)
		fmt.Println("cmd.Stderr:", string(result.Stderr))
	} else {
		fmt.Println("cmd.Stdout:", string(result.Stdout))
	}
}

func Test_run(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	// 设置要执行的命令和工作目录
	cmd := exec.Command("sh", "-c", "cd /Users/zxj && echo 11111")
	cmd.Dir = ""

	// 设置标准输出和标准错误输出
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// 执行命令
	err := cmd.Run()
	if err != nil {
		fmt.Println("命令执行失败:", err)
		fmt.Println("cmd.Stderr:", stderr.String())
		return
	} else {
		fmt.Println("cmd.Stdout:", out.String())
	}
}
