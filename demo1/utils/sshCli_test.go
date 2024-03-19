package utils

import (
	"fmt"
	"testing"
)

var sshcli = NewSSHClient("hyperchain", "hyperchain", "172.22.66.173", "22")

func Test_ssh(t *testing.T) {
	result, err := sshcli.Run("cd /home/hyperchain/jenkins/workspace/testEnv/0.0/node1 && pwd && nohup ./hyperchain  > node.log  2>&1 &")
	if err != nil {
		fmt.Print(err.Error())
	} else {
		fmt.Print(result)
	}
}

func Test_file(t *testing.T) {
	err := sshcli.Connect()
	if err != nil {
		fmt.Errorf("connect server error: %w", err)
	}
	sshcli.Upload("/Users/zxj/hyperchain2.0", "/home/hyperchain/jenkins/workspace/testEnv/node1")
}
