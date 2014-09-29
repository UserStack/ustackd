package server

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

func (server *Server) demonize() (err error) {
	chrootPath := server.Cfg.Security.Chroot
	if chrootPath != "" {
		if err = server.chroot(chrootPath); err != nil {
			return
		}
	}

	uid := server.Cfg.Security.Uid
	if uid != "" {
		if err = server.dropPrivileges(uid); err != nil {
			return
		}
	}

	err = server.checkPidFile(server.Cfg.Daemon.Pid, server.App.Name)
	if err != nil {
		return
	}
	server.writePidFile(server.Cfg.Daemon.Pid)
	return
}

func (server *Server) chroot(chrootPath string) (err error) {
	if err = syscall.Chroot(chrootPath); err != nil {
		return
	}
	err = syscall.Chdir("/")
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
		println("123")
		return
	}
	output, _ := exec.Command("ps", "-o", "command=", strconv.Itoa(pid)).Output()
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
	return strconv.Atoi(string(line))
}

func (server *Server) writePidFile(pidFile string) {
	file, err := os.Create(pidFile)
	defer file.Close()
	if err == nil {
		pid := os.Getpid()
		file.WriteString(strconv.Itoa(pid))
	}
}

func (server *Server) checkSignal(isRunning *bool, cb func() error) {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill)
	<-channel //Block until a signal is received
	os.Remove(server.Cfg.Daemon.Pid)
	*isRunning = false
	cb()
}
