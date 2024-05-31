package wakanda

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"git.pubmatic.com/PubMatic/go-common/logger"
)

var (
	commandHandler *CommandHandler
)

// send transfers the data at destFileName using SFTP configuration
//
//	StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null
//
// So there is not strict host check and if a host key changes, then there is no problem
func send(destFileName, pubProfDir string, data []byte, cfg SFTP) error {
	if commandHandler == nil {
		return errors.New("commandHandler is nil")
	}
	srcFile, err := os.CreateTemp("", destFileName)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	srcFile.Write(data)
	user := cfg.User
	password := cfg.Password
	source := srcFile.Name()
	serverIp := cfg.ServerIP
	dest := cfg.Destination + "/" + pubProfDir
	cmd := commandHandler.commandExecutor.Command()
	cmdWriter, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	cmdString := fmt.Sprintf(`
	sshpass -v -p"%s" sftp -oStrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -oBatchMode=no -b - %s@%s << !
-mkdir  %s
-chmod 777 %s
-put %s %s
-chmod 777 %s/%s
!
	`, password,
		user,
		serverIp,
		dest,                  // for mkdir
		dest,                  // for chmod
		source,                // file contains wakanda logs
		dest+"/"+destFileName, // where source file to be copied
		dest,                  // chmod file
		destFileName,          // chmod file
	)

	cmdWriter.Write([]byte(cmdString + "\n"))
	cmdWriter.Write([]byte("exit" + "\n"))

	logger.Debug("[WAKANDA] file:[%v] command:[%v]", srcFile.Name(), cmdString)
	go func() {
		defer os.Remove(srcFile.Name())
		err = cmd.Wait()
		if err != nil {
			logger.Error("SFTP_WAKANDA- '%s' SFTP Error : %s", destFileName, err.Error())
		}
	}()

	return err
}

type CommandHandler struct {
	commandExecutor commandExecutor
}

type Commands interface {
	Start() error
	StdinPipe() (io.WriteCloser, error)
	Wait() error
}

type commandExecutor interface {
	Command() Commands
}

func (h *CommandHandler) Command() Commands {
	return exec.Command("sh")
}
