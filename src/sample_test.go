package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"unsafe"
	"strconv"
)

func main() {
	child := os.Getenv(IS_CHILD)
	if len(child) > 0 {
		fmt.Println("child!")
		fmt.Println(os.Environ())
		fdStr := os.Getenv(TCP_LISTENER_KEY)
		fd, _:=strconv.Atoi(fdStr)
		fmt.Println("fd: ", fd)
		f := os.NewFile(uintptr(fd), "listen socket")
		l, _ := net.FileListener(f)
		conn, _ := l.Accept()
		fmt.Println("accepted: ", conn)
		fmt.Println(conn.(*net.TCPConn))
		tcpConn := conn.(*net.TCPConn)
		tcpConn.Write([]byte("a"))
		os.Exit(1)
	} else {
		fmt.Println("parent!")
	}

	fmt.Println("enter main!!!!")
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:9000")
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//	var name = "byebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebyebye"
	//	SetProcessName(name)
	//	fmt.Println(name)

	fmt.Println(os.Getwd())

	//connect
//	_, acceptErr := tcpListener.AcceptTCP()
//	fmt.Println("accept err: ", acceptErr)

	StartNewThisWithTcpListener(tcpListener)
	fmt.Println("restart done!")

	select {}

}

// total len = orig
// if less than orig, has orig rest
func SetProcessName(name string) error {
	fmt.Println("pointers:")
	fmt.Printf("%p\n", &os.Args)
	fmt.Printf("%p %p %p\n", &os.Args[0], &os.Args[1], &os.Args[2])
	argv0str := (*reflect.StringHeader)(unsafe.Pointer(&os.Args[0]))
	argv0 := (*[1<<30]byte)(unsafe.Pointer(argv0str.Data))[:argv0str.Len]

	n := copy(argv0, name)
	if n < len(argv0) {
		argv0[n] = 0
	}

	return nil
}
func SetProcessName2(name string) error {
	bytes := append([]byte(name), 0)
	ptr := unsafe.Pointer(&bytes[0])
	if _, _, errno := syscall.RawSyscall6(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(ptr), 0, 0, 0, 0); errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}

const (
	TCP_LISTENER_KEY = "TCP_LISTENER_FD"
	IS_CHILD         = "IS_CHILD"
)

// this method does not inherit any fd, only inherit given fd ???
func StartNewThisWithTcpListener(l *net.TCPListener) error {
	execCmd, err := exec.LookPath(os.Args[0])
	if nil != err {
		return err
	}
	wd, err := os.Getwd()
	if nil != err {
		return err
	}
	netFD := reflect.ValueOf(l).Elem().FieldByName("fd").Elem()
	fd := uintptr(netFD.FieldByName("sysfd").Int())
	inheritedFiles := append([]*os.File{os.Stdin, os.Stdout, os.Stderr},
		os.NewFile(fd, string(netFD.FieldByName("sysfile").String())))

	//start process
	_, err = os.StartProcess(execCmd, os.Args, &os.ProcAttr{
		Dir:   wd,
		Env:   append(os.Environ(), fmt.Sprintf("%s=%d", TCP_LISTENER_KEY, fd), fmt.Sprintf("%s=%s", IS_CHILD, "true")),
		Files: inheritedFiles,
	})

	if nil != err {
		return err
	}
	return nil
}

