package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// docker 				run <image> <cmd> <params>
// go run main.go run 				<cmd> <params>

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("Help!")
	}
}

func run() {
	fmt.Printf("Running %v as PID %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as PID %d\n", os.Args[2:], os.Getpid())

	cg()

	syscall.Sethostname([]byte("my-container"))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())
	syscall.Unmount("proc", 0)
}

func cg() {
	pids := "/sys/fs/cgroup/pids"
	os.Mkdir(filepath.Join(pids, "my-container"), 0755)
	must(ioutil.WriteFile(filepath.Join(pids, "my-container/pids.max"), []byte("20"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "my-container/cgroup.procs"), []byte(strconv.Itoa(os.Getgid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
