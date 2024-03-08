package console

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/euphoria-io/scope"
	"golang.org/x/crypto/ssh"

	"euphoria.leet.nu/heim/cluster"
	"euphoria.leet.nu/heim/console"
	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/logging"
	"euphoria.leet.nu/heim/proto/security"
)

type Controller struct {
	m        sync.Mutex
	closed   bool
	listener net.Listener
	config   *ssh.ServerConfig
	backend  proto.Backend
	kms      security.KMS
	cluster  cluster.Cluster
	ctx      scope.Context

	// TODO: key ssh.PublicKey
	authorizedKeys []ssh.PublicKey
}

func NewController(heim *proto.Heim, addr string) (*Controller, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %s", addr, err)
	}

	ctrl := &Controller{
		listener: listener,
		backend:  heim.Backend,
		kms:      heim.KMS,
		cluster:  heim.Cluster,
		ctx:      heim.Context,
	}

	ctrl.config = &ssh.ServerConfig{
		PublicKeyCallback: ctrl.authorizeKey,
	}

	return ctrl, nil
}

func (ctrl *Controller) Close() error {
	ctrl.m.Lock()
	defer ctrl.m.Unlock()
	if ctrl.closed {
		return nil
	}
	ctrl.closed = true
	return ctrl.listener.Close()
}

func (ctrl *Controller) authorizeKey(conn ssh.ConnMetadata, key ssh.PublicKey) (
	*ssh.Permissions, error) {

	marshaledKey := key.Marshal()
	for _, authorizedKey := range ctrl.authorizedKeys {
		if bytes.Compare(authorizedKey.Marshal(), marshaledKey) == 0 {
			return &ssh.Permissions{}, nil
		}
	}

	nodes, err := ctrl.cluster.GetDir("console/authorized_keys")
	if err != nil {
		if err == cluster.ErrNotFound {
			return nil, fmt.Errorf("unauthorized")
		}
		return nil, err
	}

	for path, value := range nodes {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(value))
		if err != nil {
			fmt.Printf("bad authorized key from etcd: %s: %s\n", path, err)
		}
		if bytes.Compare(key.Marshal(), marshaledKey) == 0 {
			return &ssh.Permissions{}, nil
		}
	}

	return nil, fmt.Errorf("unauthorized")
}

func (ctrl *Controller) AddHostKeyFromCluster(host string) error {
	generate := func() (string, error) {
		// Generate an ECDSA key.
		key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return "", err
		}
		derBytes, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return "", err
		}
		w := &bytes.Buffer{}
		if err := pem.Encode(w, &pem.Block{Type: "EC PRIVATE KEY", Bytes: derBytes}); err != nil {
			return "", err
		}
		return w.String(), nil
	}
	pemString, err := ctrl.cluster.GetValueWithDefault(fmt.Sprintf("console/%s", host), generate)
	if err != nil {
		return fmt.Errorf("failed to get/generate host key: %s", err)
	}

	signer, err := ssh.ParsePrivateKey([]byte(pemString))
	if err != nil {
		return fmt.Errorf("failed to parse host key: %s", err)
	}

	ctrl.config.AddHostKey(signer)
	return nil
}

func (ctrl *Controller) AddHostKey(path string) error {
	pemBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	key, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return err
	}

	ctrl.config.AddHostKey(key)
	return nil
}

func (ctrl *Controller) AddAuthorizedKeys(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	curLine := 1

	for {
		startLine := curLine
		buf := &bytes.Buffer{}

		for {
			line, isPrefix, err := r.ReadLine()
			if err != nil && err != io.EOF {
				return err
			}
			buf.Write(line)
			curLine++
			if !isPrefix {
				break
			}
		}

		line := bytes.TrimSpace(buf.Bytes())
		if len(line) == 0 {
			break
		}

		key, _, _, _, err := ssh.ParseAuthorizedKey(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			fmt.Printf("%s:%d: not a public key: %s\n", path, startLine, err)
			continue
		}

		ctrl.authorizedKeys = append(ctrl.authorizedKeys, key)
	}

	return nil
}

func (ctrl *Controller) Serve() {
	defer ctrl.ctx.WaitGroup().Done()

	conns := make(chan net.Conn)
	go func() {
		for ctrl.ctx.Err() == nil {
			conn, err := ctrl.listener.Accept()
			if err != nil {
				ctrl.m.Lock()
				if !ctrl.closed {
					ctrl.ctx.Terminate(fmt.Errorf("console accept: %s", err))
				}
				ctrl.m.Unlock()
				return
			}
			conns <- conn
		}
	}()

	for {
		select {
		case <-ctrl.ctx.Done():
			return
		case conn := <-conns:
			ctrl.ctx.WaitGroup().Add(1)
			go ctrl.interact(ctrl.ctx.Fork(), conn)
		}
	}
}

func (ctrl *Controller) interact(ctx scope.Context, conn net.Conn) {
	defer ctx.WaitGroup().Done()

	_, nchs, reqs, err := ssh.NewServerConn(conn, ctrl.config)
	if err != nil {
		return
	}

	go ssh.DiscardRequests(reqs)

	for nch := range nchs {
		if nch.ChannelType() != "session" {
			nch.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		ch, reqs, err := nch.Accept()
		if err != nil {
			return
		}
		go ctrl.filterClientRequests(reqs)
		go ctrl.terminal(ctx, ch)
	}
}

func (ctrl *Controller) filterClientRequests(reqs <-chan *ssh.Request) {
	for req := range reqs {
		switch req.Type {
		case "shell":
			req.Reply(len(req.Payload) == 0, nil)
		case "pty-req":
			req.Reply(true, nil)
		default:
			req.Reply(false, nil)
		}
	}
}

func (ctrl *Controller) terminal(ctx scope.Context, ch ssh.Channel) {
	defer ch.Close()

	subctx := logging.LoggingContext(ctx.Fork(), os.Stdout,
		fmt.Sprintf("[console %p] ", ch))

	cli := newCLI(ctrl)
	con := console.NewDefaultConsole(ch).WithContext(subctx)
	defer con.Close()
	console.LaunchCLI(con, cli.CLI, nil)
}
