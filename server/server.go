package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

var (
	users       = make(map[string]*net.Conn)
	normalStyle = ansi.NewStyle(ansi.AttrGreenForegroundColor)
	errStyle    = ansi.NewStyle(ansi.AttrRedForegroundColor)
)

func Listen(port string) {
	l, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(normalStyle.Styled("Server start Listen on " + l.Addr().String()))
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print(errStyle.Styled(err.Error()))
		}
		go login(conn)
	}
}

func login(conn net.Conn) {
	input := make([]byte, 10)

	for {
		n, err := conn.Read(input)
		if err != nil {
			if err != io.EOF {
				conn.Write([]byte{3})
				log.Print(errStyle.Styled(err.Error()))
				continue
			}
		}

		username := string(input[:n-1])
		if n > 0 {
			_, ok := users[username]
			if ok {
				conn.Write([]byte{2})
				continue
			}
		}

		users[username] = &conn
		conn.Write([]byte{1})
		log.Println(normalStyle.Styled("+++ Added user " + username))
		go messageReader(username)
		return
	}
}

func messageReader(username string) {
	conn := *users[username]
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Print(errStyle.Styled(username + " have Error: " + err.Error()))
			}
			log.Println(normalStyle.Styled("--- Deleting user " + username))
			delete(users, username)
			return
		}

		msg = strings.TrimRight(msg, "\n")
		go hub(username, msg)
	}
}

func hub(username, msg string) {
	fullmsg := "\x1b[32m" + username + ":\x1b[0m " + msg
	for resever := range users {
		if resever == username {
			continue
		}
		fmt.Println("sending to ", resever)
		_, err := (*users[resever]).Write([]byte(fullmsg))
		if err != nil {
			log.Print(errStyle.Styled(resever + " have Error: " + err.Error()))
			log.Println(normalStyle.Styled("Deleting user " + resever))
			delete(users, resever)
		}
	}
}
