package main

import (
	"fmt"
	"log"
	"os"

	"github.com/skorobogatov/input"
	"golang.org/x/crypto/ssh"
)

func main() {

	username := "alex"
	hostname := "0.0.0.0"
	port := "2200"
	client, session, err := connectToHost(username, hostname+":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Shell()
	if err != nil {
		log.Fatal(err)
	}

	for {
		cmd := input.Gets()
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Fatal(err)
		}
		if cmd == "exit" {
			session.Close()
			break
		}
	}
	client.Close()

}


func connectToHost(user, host string) (*ssh.Client, *ssh.Session, error) {
	var pass string
	fmt.Print("Password: ")
	fmt.Scanf("%s\n", &pass)
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(pass)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}
