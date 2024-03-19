package utils

import (
	"fmt"
	"io"
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
	if c.sftpClient == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}
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

func (c Cli) Download(remotePath, localPath string) error {
	if c.sftpClient == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}
	remoteFileInfo, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("无法获取远程文件或文件夹信息：%s", err)
	}
	if remoteFileInfo.IsDir() {
		return c.downloadDir(remotePath, localPath)
	} else {
		return c.downloadFile(remotePath, localPath)
	}
}

// localPath是个文件夹
func (c Cli) downloadFile(remotePath, localPath string) error {
	remoteFile, err := c.sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("无法打开远程文件：%s", err)
	}
	defer remoteFile.Close()

	var localFileName = path.Base(remotePath)
	localFile, err := os.Create(path.Join(localPath, localFileName))
	if err != nil {
		return fmt.Errorf("无法创建本地文件：%s", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("文件下载失败：%s", err)
	}
	return nil
}

func (c Cli) downloadDir(remotePath, localPath string) error {
	// 获取远程文件夹中的文件列表
	remoteFiles, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return fmt.Errorf("无法读取远程文件夹中的文件：%s", err)
	}
	// 递归下载文件夹中的文件
	for _, remoteFile := range remoteFiles {
		remoteFilePath := path.Join(remotePath, remoteFile.Name())
		localFilePath := path.Join(localPath, remoteFile.Name())

		if remoteFile.IsDir() {
			os.Mkdir(remoteFilePath, 0755)
			err := c.downloadDir(remoteFilePath, localFilePath)
			if err != nil {
				return err
			}
		} else {
			err := c.downloadFile(remoteFilePath, localPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// remotePath是个文件夹
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
			c.uploadFile(localFilePath, remotePath)
		}
	}
	return nil
}
