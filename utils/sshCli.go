package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Cli struct {
	user       string
	pwd        string
	ip         string
	port       string
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func NewSSHClient(user, pwd, ip, port string) Cli {
	return Cli{
		user: user,
		pwd:  port,
		ip:   ip,
		port: port,
	}
}

// 不使用 HostKey， 使用密码
func (c *Cli) getConfig_nokey() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.pwd),
		},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

func (c *Cli) Connect() error {
	config := c.getConfig_nokey()
	client, err := ssh.Dial("tcp", c.ip+":"+c.port, config)
	if err != nil {
		return fmt.Errorf("connect server error: %w", err)
	}
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("new sftp client error: %w", err)
	}
	c.sshClient = client
	c.sftpClient = sftp
	return nil
}

func (c Cli) Run(cmd string) (string, error) {
	if c.sshClient == nil {
		if err := c.Connect(); err != nil {
			return "", err
		}
	}
	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("create new session error: %w", err)
	}
	defer session.Close()
	buf, err := session.CombinedOutput(cmd)
	return string(buf), err
}

func (c Cli) Upload(localPath, remotePath string) error {
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("无法获取本地路径信息：%s", err)
	}

	if localInfo.IsDir() {
		return c.uploadDir(localPath, remotePath)
	} else {
		return c.uploadFile(localPath, remotePath)
	}
}

func (c Cli) uploadFile(localPath, remotePath string) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("无法打开本地文件：%s", err)
	}
	defer localFile.Close()

	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("无法创建远程文件：%s", err)
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("文件上传失败：%s", err)
	}
	return nil
}

func (c Cli) uploadDir(localPath, remotePath string) error {
	localFiles, err := filepath.Glob(filepath.Join(localPath, "*"))
	if err != nil {
		return fmt.Errorf("无法读取本地文件夹：%s", err)
	}
	for _, file := range localFiles {
		localInfo, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("无法获取本地路径信息：%s", err)
		}

		remoteFilePath := filepath.Join(remotePath, localInfo.Name())
		if localInfo.IsDir() {
			err = c.uploadDir(file, remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			err = c.uploadFile(file, remoteFilePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
