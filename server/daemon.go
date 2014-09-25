package server

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

func (server *Server) Demonize() (err error) {
	uid := server.Cfg.Security.Uid
	if uid != "" {
		if err = server.dropPrivileges(uid); err != nil {
			server.Logger.Println(err)
			return
		}
	}

	pidFile := server.Cfg.Daemon.Pid_Path + "/" + server.AppName + ".pid"
	err = server.checkPidFile(pidFile, server.AppName)
	if err != nil {
		server.Logger.Println(err.Error())
		return
	}
	server.writePidFile(pidFile)
	return
}

func (server *Server) dropPrivileges(username string) (err error) {
	usr, err := user.Lookup(username)
	if err != nil {
		return
	}
	if usr == nil {
		return fmt.Errorf("User %s does not exist on system.", username)
	}
	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		return
	}
	gid, err := strconv.Atoi(usr.Gid)
	if err != nil {
		return
	}

	if syscall.Getuid() == 0 {
		_, err = syscall.Getgroups()
		if err != nil {
			return
		}
		err = syscall.Setgroups([]int{gid})
		if err != nil {
			return
		}
		_, err = syscall.Getgroups()
		if err != nil {
			return
		}
		err = syscall.Setregid(gid, gid)
		if err != nil {
			return
		}
		err = syscall.Setreuid(uid, uid)
		if err != nil {
			return
		}
		if syscall.Getuid() != 0 {
			server.Logger.Println("Privileges succesfully dropped.")
		}
	}
	return
}

func (server *Server) checkPidFile(pidFile, appname string) (err error) {
	if _, ferr := os.Stat(pidFile); ferr != nil {
		return
	}
	pid, err := server.readPidFile(pidFile)
	if err != nil {
		return
	}
	output, err := exec.Command("ps", "-o", "command=", strconv.Itoa(pid)).Output()
	if err != nil {
		return
	}
	if strings.Contains(string(output), appname) {
		err = fmt.Errorf("Running %s found with PID: %d", appname, pid)
		return
	}
	os.Remove(pidFile)
	return
}

func (server *Server) readPidFile(pidFile string) (pid int, err error) {
	file, err := os.Open(pidFile)
	defer file.Close()
	if err != nil {
		return
	}
	reader := bufio.NewReaderSize(file, 5)
	line, _, err := reader.ReadLine()
	if err != nil {
		return
	}
	pid, err = strconv.Atoi(string(line))
	return
}

func (server *Server) writePidFile(pidFile string) {
	file, err := os.Create(pidFile)
	defer file.Close()
	if err == nil {
		pid := os.Getpid()
		file.WriteString(strconv.Itoa(pid))
	}
}
