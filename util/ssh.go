package util

import (
	ssh "golang.org/x/crypto/ssh"
	"fmt"
	log "gatoor/orca/base/log"
	"github.com/Sirupsen/logrus"
	"io"
	"time"
	"io/ioutil"
)

const (
	SSH_PEM_PATH = "/orca/config/orca_test_us.pem"
	CONNECT_RETRY_AMOUNT = 15
)

func Connect(sshUser string, hostAndPort string) (*ssh.Client, string) {
	addr := sshUser + "@" + hostAndPort
	var SSHLogger = log.LoggerWithField(log.LoggerWithField(log.AuditLogger, "Type", "ssh"), "target", addr)

	pemBytes, err := ioutil.ReadFile(SSH_PEM_PATH)
	if err != nil {
		SSHLogger.Errorf("PEM file read failed: %s", err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		SSHLogger.Errorf("PEM file parse failed: %s", err)
	}

	SSHLogger.Info("Connecting...")
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		Timeout: time.Second * 3,
	}

	var connection *ssh.Client
	for i := 1; i <= CONNECT_RETRY_AMOUNT; i++ {
		connection, err = ssh.Dial("tcp", hostAndPort, sshConfig)

		if err != nil {
			if i == CONNECT_RETRY_AMOUNT {
				SSHLogger.Error(fmt.Sprintf("Failed to dial: %v aborting", err))
				return nil, ""
			}
			SSHLogger.Error(fmt.Sprintf("Failed to dial, it was try %d: %v retrying...", i, err))
			time.Sleep(time.Duration(5 * time.Second))
			continue
		} else {
			return connection, addr
		}
	}
	return nil, ""
}

func accquireSession(connection *ssh.Client, SSHLogger *logrus.Entry, stdWriter io.Writer) *ssh.Session {
	session, err := connection.NewSession()
	if err != nil {
		SSHLogger.Error(fmt.Sprintf("Failed to create session: %v", err))
		return nil
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		SSHLogger.Error(fmt.Sprintf("Request for pseudo terminal failed: %v", err))
		return nil
	}
	session.Stdout = stdWriter
	session.Stderr = stdWriter
	SSHLogger.Info("Created session.")
	return session
}

func ExecuteSshCommand(conn *ssh.Client, addr string, cmd string) bool {
	var SSHLogger = log.LoggerWithField(log.LoggerWithField(log.AuditLogger, "Type", "ssh"), "target", addr)
	SSHLogger.Info(fmt.Sprintf("Executing command: [%s]", cmd))
	stdWriter := SSHLogger.Logger.Writer()
	session := accquireSession(conn, SSHLogger, stdWriter)
	defer session.Close()
	defer stdWriter.Close()
	err := session.Run(cmd)
	if err != nil {
		SSHLogger.Error(fmt.Sprintf("Command failed - %s", err))
		return false
	}
	return true
}