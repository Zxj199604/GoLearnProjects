package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
		pwd:  pwd,
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
		Timeout:         30 * time.Minute,
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

func (c Cli) uploadFile(localFilePath, remotePath string) error {
	fmt.Println(localFilePath + " copy file to remote server start!")
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("os.Open error : %s", localFilePath)
	}
	defer srcFile.Close()
	var remoteFileName = path.Base(localFilePath)
	dstFile, err := c.sftpClient.Create(path.Join(remotePath, remoteFileName))
	if err != nil {
		return fmt.Errorf("sftpClient.Create error : %s", path.Join(remotePath, remoteFileName))
	}
	defer dstFile.Close()
	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("ReadAll error : %s", localFilePath)
	}
	dstFile.Write(ff)
	fmt.Println(localFilePath + " copy file to remote server finished!")
	return nil
}

func (c Cli) uploadDir(localPath, remotePath string) error {
	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("read dir list fail:%s ", err)
	}
	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			c.sftpClient.Mkdir(remoteFilePath)
			c.uploadDir(localFilePath, remoteFilePath)
		} else {
			c.uploadFile(path.Join(localPath, backupDir.Name()), remotePath)
		}
	}
	return nil
}
