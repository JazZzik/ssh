package main
 
import (
    "fmt"
    "strings"
    "io/ioutil"
    "log"
    "net"
    "os/exec"
    "os"
 
    "golang.org/x/crypto/ssh"
)

func getWP() ([]byte){
	path, err := os.Getwd()
	if err != nil {
	    log.Println(err)
	}
	pat := strings.Split(path, "/")
	pat = pat[6:]
	pa := strings.Join(pat[:], "/")
	p := []byte(pa)
	b := []byte(">")
	p = append(p, b...)
	return p
}
 
func passwordCallback(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	if (c.User() == "alex" && string(pass) == "123") {
		return nil, nil
	}
	return nil, fmt.Errorf("Password rejected for %q", c.User())
}
 
func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}
 
func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	
	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Fatalf("Could not accept channel: %v", err)
	}
	nRequests := make(chan *ssh.Request, 15)
	for req := range requests {
		nRequests <- req
		if req.Type == "shell" || req.Type == "exec" {
			break
		}
	}
	
	go func(in <-chan *ssh.Request) {
		for req := range in {
			req.Reply(req.Type == "shell" || req.Type == "exec", nil)
		}
	} (nRequests)

		
	a := make([]byte, 32)
	defer channel.Close()
	for {
		channel.Read(a)
		cmd := string(a)
		i := strings.Index(cmd, "\n")
		cmd = cmd[:i]
		if cmd == "exit" {
			channel.Close()
			break
		}
		command := strings.Split(cmd, " ")
		if command[0] == "cd"{
			os.Chdir(command[1])
			p := getWP()
			channel.Write(p)
		} else {
			res, err := exec.Command(command[0], command[1:]...).Output()
			if err != nil {
				log.Print(command, err)
				channel.Write([]byte(err.Error() + "\n"))
			}
			p := getWP()
			res = append(p, res...)
			channel.Write(res)
		}
	}

}
 
 
func main() {
	config := &ssh.ServerConfig{
		PasswordCallback: passwordCallback,
	}

	privateBytes, err := ioutil.ReadFile("/home/alex/.ssh/id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key (./id_rsa)")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)


	listener, err := net.Listen("tcp", "0.0.0.0:2200")
	if err != nil {
		log.Fatalf("Failed to listen on 2200 (%s)", err)
	}

	log.Print("Listening on 2200...")
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
	    		log.Printf("Failed to accept incoming connection (%s)", err)
	    		continue
		}
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
}

