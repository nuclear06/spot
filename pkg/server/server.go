package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	. "github.com/ssh-honeypot/pkg/config"
	. "github.com/ssh-honeypot/pkg/logging"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"net"
	"os"
)

var (
	PublicKeyAccept bool
	NoAuthAccept    bool
	PasswordAccept  bool
)

func InitSSH(config Config) error {
	sshConf := &ssh.ServerConfig{
		NoClientAuth:  config.NoClientAuth,
		MaxAuthTries:  config.MaxAuthTries,
		ServerVersion: config.ServerVersion,
		BannerCallback: func(conn ssh.ConnMetadata) string {
			return config.Banner + "\r\n"
		},
	}
	PasswordAccept, PublicKeyAccept, NoAuthAccept = config.PasswordAuth.Accept, config.PublicKeyAuth.Accept, config.NoClientAuth
	if config.PasswordAuth.Enable {
		sshConf.PasswordCallback = pwdCallback
		LogDebug(SysInit, "enable password auth")
	}
	if config.PublicKeyAuth.Enable {
		sshConf.PublicKeyCallback = pubkeyCallback
		LogDebug(SysInit, "enable public key auth")
	}
	if config.NoClientAuth {
		sshConf.NoClientAuthCallback = noAuthCallback
		LogDebug(SysInit, "enable no auth connection")
	}

	err := Listen(config, sshConf)
	return err
}
func pwdCallback(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	cV := c.ClientVersion()
	if PasswordAccept {
		SSHInfo(PasswordAuth, c.RemoteAddr().String(), fmt.Sprintf("permit client: %s", cV), logrus.Fields{"user": c.User(), "password": string(pass)})
		return &ssh.Permissions{}, nil
	}
	SSHInfo(PasswordAuth, c.RemoteAddr().String(), fmt.Sprintf("reject client: %s", cV), logrus.Fields{"user": c.User(), "password": string(pass)})
	return nil, fmt.Errorf("password rejected for [%s]", c.User())
}
func pubkeyCallback(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	if PublicKeyAccept {
		SSHInfo(PublicKeyAuth, c.RemoteAddr().String(), fmt.Sprintf("permit client: %s", c.ClientVersion()), logrus.Fields{"user": c.User(), "key_finger": ssh.FingerprintSHA256(pubKey)})
		return &ssh.Permissions{}, nil
	}
	SSHInfo(PublicKeyAuth, c.RemoteAddr().String(), fmt.Sprintf("reject client: %s", c.ClientVersion()), logrus.Fields{"user": c.User(), "key_finger": ssh.FingerprintSHA256(pubKey)})
	return nil, fmt.Errorf("unknown public key for [%s]", c.User())
}
func noAuthCallback(c ssh.ConnMetadata) (*ssh.Permissions, error) {
	if NoAuthAccept {
		SSHInfo(NoAuth, c.RemoteAddr().String(), fmt.Sprintf("permit client: %s", c.ClientVersion()), logrus.Fields{"user": c.User()})
		return &ssh.Permissions{}, nil
	}
	return nil, nil
}

func handleChannel(addr string, chans <-chan ssh.NewChannel) {
	// Service the incoming channel.
	for newChannel := range chans {
		if t := newChannel.ChannelType(); t != "session" {
			err := newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			SSHDebug(SysRunning, addr, fmt.Sprintf("reject channel, type: %v", t), nil)
			if err != nil {
				SSHWarn(SysRunning, addr, fmt.Sprintf("reject channel failed (%v)", err), nil)
			}
			continue
		}

		ch, requests, err := newChannel.Accept()
		if err != nil {
			SSHWarn(SysRunning, addr, fmt.Sprintf("could not accept channel (%s)", err), nil)
			continue
		}

		for req := range requests {
			resp := true
			switch req.Type {
			case "exec":
				var execReq struct {
					Command string
				}
				if err := ssh.Unmarshal(req.Payload, &execReq); err != nil {
					SSHWarn(SshExec, addr, fmt.Sprintf("exec payload unmarshal failed (%v)", err), nil)
				}

				resp = false
				err := ch.Close()
				if err != nil {
					return
				}
				SSHInfo(SshExec, addr, fmt.Sprintf("exec:[%s]", execReq.Command), nil)
			case "shell":
				go func(c ssh.Channel, resp *bool) {
					SSHDebug(SshShell, addr, "allocate shell", nil)
					term := terminal.NewTerminal(c, "$ ")
					for {
						line, err := term.ReadLine()
						if err != nil {
							if err != io.EOF {
								SSHWarn(SshShell, addr, fmt.Sprintf("terminal read failed (%v)", err), nil)
							}
							// err is EOF, usually means the client has exited
							SSHDebug(SshShell, addr, "shell exit", nil)

							*resp = false
							err := c.Close()
							if err != nil {
								return
							}
							return
						}
						SSHInfo(SshShell, addr, fmt.Sprintf("shell received:[%s]", line), nil)
					}
				}(ch, &resp)
			case "pty-req":
				SSHDebug(SshPtyReq, addr, "pty request", nil)
			case "env":
				var envReq struct {
					Name  string
					Value string
				}
				if err := ssh.Unmarshal(req.Payload, &envReq); err != nil {
					SSHWarn(SshEnv, addr, fmt.Sprintf("env payload unmarshal failed (%v)", err), nil)
				}
				SSHDebug(SshEnv, addr, "env request", logrus.Fields{"name": envReq.Name, "value": envReq.Value})
			default:
				SSHDebug(SshOther, addr, fmt.Sprintf("reject request, type: %v", req.Type), nil)
			}
			if resp {
				err := req.Reply(true, nil)
				if err != nil {
					SSHWarn(SysRunning, addr, fmt.Sprintf("request reply failed (%v)", err), nil)
				}
			}
		}
	}
}

func Listen(conf Config, serverConf *ssh.ServerConfig) error {
	defer func() {
		if err := recover(); err != nil {
			LogFatal(SysInit, fmt.Sprintf("Failed to listen on %s, err: %v", conf.Addr, err))
			os.Exit(1)
		}
	}()

	if hostKeys := conf.HostKeys; hostKeys == nil || len(hostKeys) == 0 {
		LogFatal(SysInit, "Must specify at least one host key")
		os.Exit(1)
	}
	// load host keys
	for _, hostKey := range conf.HostKeys {
		// auto generate host key if not exist
		if _, err := os.Stat(hostKey); os.IsNotExist(err) {
			LogError(SysInit, fmt.Sprintf("host key file not exist: %s", hostKey))
			LogInfo(SysInit, fmt.Sprintf("Generating host key: %s", hostKey))
			err = genPrivateKey(hostKey)
			if err != nil {
				LogFatal(SysInit, fmt.Sprintf("Failed to generate host keys, err: %v", err))
				os.Exit(1)
			}
		}

		prvKey, err := os.ReadFile(hostKey)
		if err != nil {
			return err
		}
		private, err := ssh.ParsePrivateKey(prvKey)
		if err != nil {
			return err
		}
		signer, err := ssh.NewSignerWithAlgorithms(private.(ssh.AlgorithmSigner), []string{ssh.KeyAlgoRSASHA256, ssh.KeyAlgoRSASHA512})
		if err != nil {
			return err
		}
		pK := signer.PublicKey()
		LogDebug(SysInit, fmt.Sprintf("load host key: [%s] = %s|%s", hostKey, pK.Type(), ssh.FingerprintSHA256(pK)))
		serverConf.AddHostKey(signer)
	}

	listener, err := net.Listen("tcp", conf.Addr)
	LogInfo(SysInit, "SSH server is listening on "+conf.Addr)
	if err != nil {
		return err
	}
	for {
		nConn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func(Conn net.Conn) {
			// handshake
			conn, chans, reqs, err := ssh.NewServerConn(Conn, serverConf)
			if err != nil {
				// usually means ssh client has exited
				SSHDebug(SysRunning, "", fmt.Sprintf("handshake failed: %v", err), nil)
				return
			}
			addr := conn.RemoteAddr().String()
			go ssh.DiscardRequests(reqs)
			go handleChannel(addr, chans)
		}(nConn)
	}
	return nil
}
