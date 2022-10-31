package cluster

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/google/uuid"
)

type Node struct {
	ID    string
	Addr  string
	proc  *exec.Cmd
	proxy *toxiproxy.Proxy
}

func NewNode(portAllocator *PortAllocator, toxiproxyClient *toxiproxy.Client) (*Node, error) {
	id := uuid.New().String()

	listenAddr := fmt.Sprintf("127.0.0.1:%d", portAllocator.Take())
	proxyAddr := fmt.Sprintf("127.0.0.1:%d", portAllocator.Take())

	proxyID := fmt.Sprintf("wombat_%s", id)
	proxy, err := toxiproxyClient.CreateProxy(proxyID, proxyAddr, listenAddr)
	if err != nil {
		return nil, err
	}

	gossipAddr := fmt.Sprintf("127.0.0.1:%d", portAllocator.Take())

	stdoutLogger, err := createStdoutLogger(id)
	if err != nil {
		return nil, err
	}

	stderrLogger, err := createStderrLogger(id)
	if err != nil {
		return nil, err
	}

	newpath := filepath.Join(".", "out")
	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	p, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	p = path.Join(p, "./out/wombat")

	cmd := exec.Command("/usr/local/bin/go", "build", "-o", p, "cmd/wombat/main.go")
	cmd.Dir = "../../service"
	cmd.Stdin = nil
	cmd.Stdout = stdoutLogger
	cmd.Stderr = stderrLogger
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	params := map[string]string{
		"--addr":        listenAddr,
		"--gossip.addr": gossipAddr,
		"--gossip.peer": id,
	}

	args := []string{}
	for k, v := range params {
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}

	cmd = exec.Command(p, args...)
	cmd.Stdin = nil
	cmd.Stdout = stdoutLogger
	cmd.Stderr = stderrLogger
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Node{
		ID:    id,
		Addr:  proxyAddr,
		proc:  cmd,
		proxy: proxy,
	}, nil
}

func (n *Node) Kill() error {
	if n.proc.Process == nil {
		return nil
	}
	return n.proc.Process.Kill()
}

func (n *Node) Signal(sig os.Signal) error {
	if n.proc.Process == nil {
		return nil
	}
	return n.proc.Process.Signal(sig)
}

func (n *Node) Wait() error {
	return n.proc.Wait()
}

func (n *Node) Shutdown() error {
	if err := n.proxy.Delete(); err != nil {
		return err
	}
	if err := n.Kill(); err != nil {
		return err
	}
	n.Wait()
	return nil
}

func createStdoutLogger(id string) (io.Writer, error) {
	if err := createLogDir(id); err != nil {
		return nil, err
	}
	f, err := os.Create("out/" + id + "/out.log")
	if err != nil {
		return nil, err
	}
	return f, nil
}

func createStderrLogger(id string) (io.Writer, error) {
	if err := createLogDir(id); err != nil {
		return nil, err
	}
	f, err := os.Create("out/" + id + "/err.log")
	if err != nil {
		return nil, err
	}
	return f, nil
}

func createLogDir(id string) error {
	err := os.MkdirAll("out/"+id, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
