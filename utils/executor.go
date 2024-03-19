package utils

import (
	"bytes"
	"os/exec"
)

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

func ExecuteCommand(dir, cmd string) CommandResult {
	// 创建命令对象
	command := exec.Command("sh", "-c", cmd)
	command.Dir = dir

	// 执行命令并捕获输出
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()

	// 返回结果
	return CommandResult{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
		Err:    err,
	}
}
