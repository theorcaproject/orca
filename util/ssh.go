package util

import (
	ssh "golang.org/x/crypto/ssh"
	"fmt"
	log "gatoor/orca/base/log"
	"github.com/Sirupsen/logrus"
	"io"
	"time"
)

func Connect(sshUser string, hostAndPort string) (*ssh.Client, string) {
	addr := sshUser + "@" + hostAndPort
	var SSHLogger = log.LoggerWithField(log.LoggerWithField(log.AuditLogger, "Type", "ssh"), "target", addr)

	SSHLogger.Info("Connecting...")
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password("orca"),
		},
		Timeout: time.Second * 3,
	}

	connection, err := ssh.Dial("tcp", hostAndPort, sshConfig)

	if err != nil {
		SSHLogger.Error(fmt.Sprintf("Failed to dial: %v", err))
		return nil, ""
	}
	return connection, addr
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