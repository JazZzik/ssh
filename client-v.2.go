package main

import (
	"fmt"
	"log"
	"golang.org/x/crypto/ssh"
)

var (
	username = "alex"
	hostname = "0.0.0.0"
	port = "2200"
	password = "123"
)


func serve() {

	
	client, session, err := connectToHost(username, hostname+":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = session.Shell()
	if err != nil {
		log.Fatal(err)
	}

	for {
		cmd := <-querCh
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Fatal(err)
		}
		
		if cmd == "exit" {
			session.Close()
			break
		}
		a := make([]byte, 1000)
		num, _ := stdout.Read(a)

		log.Println("ssh client: sending responce: ", cmd)
		responceCh <- string(a[:num])
	}
	client.Close()

}


func connectToHost(user, host string) (*ssh.Client, *ssh.Session, error) {
	var pass string
	fmt.Printf("Enter Password for %s: ", username)
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
