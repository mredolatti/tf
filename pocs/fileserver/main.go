package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type context struct {
	debugStream io.Writer
	readOnly    bool
}

// Based on example server code from golang.org/x/crypto/ssh and server_standalone
func main() {

	var debugStderr bool
	ctx := context{debugStream: ioutil.Discard}
	flag.BoolVar(&ctx.readOnly, "R", false, "read-only server")
	flag.BoolVar(&debugStderr, "e", false, "debug to stderr")
	flag.Parse()

	if debugStderr {
		ctx.debugStream = os.Stderr
	}

	knownKeys := map[string]ssh.PublicKey{
		"martin": pubKeyOrDie("./keys/client/id_ecdsa.pub"),
	}

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			knownKeyForUser, ok := knownKeys[conn.User()]
			if !ok {
				return nil, fmt.Errorf("unknown user %s", conn.User())
			}

			if !bytes.Equal(key.Marshal(), knownKeyForUser.Marshal()) {
				return nil, fmt.Errorf("invalid key for user %s", conn.User())
			}
			return nil, nil
		},
	}

	privateBytes, err := ioutil.ReadFile("./keys/server/id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key", err)
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		log.Fatal("failed to listen for connection", err)
	}
	fmt.Printf("Listening on %v\n", listener.Addr())

	for {

		nConn, err := listener.Accept()
		if err != nil {
			log.Fatal("failed to accept incoming connection", err)
		}
		go handleConnection(nConn, config, ctx)
	}

}

func handleConnection(nConn net.Conn, config *ssh.ServerConfig, ctx context) {
	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Fatal("failed to handshake", err)
	}
	fmt.Fprintf(ctx.debugStream, "SSH server established\n")

	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of an SFTP session, this is "subsystem"
		// with a payload string of "<length=4>sftp"
		fmt.Fprintf(ctx.debugStream, "Incoming channel: %s\n", newChannel.ChannelType())
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			fmt.Fprintf(ctx.debugStream, "Unknown channel type: %s\n", newChannel.ChannelType())
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatal("could not accept channel.", err)
		}
		fmt.Fprintf(ctx.debugStream, "Channel accepted\n")

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "subsystem" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				fmt.Fprintf(ctx.debugStream, "Request: %v\n", req.Type)
				ok := false
				switch req.Type {
				case "subsystem":
					fmt.Fprintf(ctx.debugStream, "Subsystem: %s\n", req.Payload[4:])
					if string(req.Payload[4:]) == "sftp" {
						ok = true
					}
				}
				fmt.Fprintf(ctx.debugStream, " - accepted: %v\n", ok)
				req.Reply(ok, nil)
			}
		}(requests)

		serverOptions := []sftp.ServerOption{
			sftp.WithDebug(ctx.debugStream),
		}

		if ctx.readOnly {
			serverOptions = append(serverOptions, sftp.ReadOnly())
			fmt.Fprintf(ctx.debugStream, "Read-only server\n")
		} else {
			fmt.Fprintf(ctx.debugStream, "Read write server\n")
		}

		server, err := sftp.NewServer(
			channel,
			serverOptions...,
		)
		if err != nil {
			log.Fatal(err)
		}
		if err := server.Serve(); err == io.EOF {
			server.Close()
			log.Print("sftp client exited session.")
		} else if err != nil {
			log.Fatal("sftp server completed with error:", err)
		}
	}
}

func pubKeyOrDie(fn string) ssh.PublicKey {
	bytes, err := ioutil.ReadFile(fn)
	if err != nil {
		panic("error reading pub key file: " + err.Error())
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(bytes)
	if err != nil {
		panic("error parsing pub key : " + err.Error())
	}

	return pubKey
}
