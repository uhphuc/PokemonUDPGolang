package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter your username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)
		_, err = conn.Write([]byte("@join " + username))
		if err != nil {
			fmt.Println("Error joining chat:", err)
			return
		}
		buffer := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error read from UDP: ", err)
			return
		}
		if string(buffer[:n]) == "Invalid" {
			fmt.Println("Username had used, please set a other username!")
		} else { 
			fmt.Println(string(buffer[:n]))
			break 
		}
	}
	for {
		fmt.Print("Choose your begining pokemon\n1. Bulbasaur\n2. Charmander\n3. Squirtle\nYour choice(input 1, 2, or 3): ")
		choose, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error to read the choosen pokemon: ", err)
			return
		}
		choose = strings.TrimSpace(choose)
		conn.Write([]byte("@choose "+choose))
		buffer := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error read from UDP: ", err)
			return
		}
		if string(buffer[:n]) == "Cannot" {
			fmt.Println("Wrong Syntax! Try again")
		} else { 
			fmt.Print("You choose correctly! Now try some option!\nUsages:\n"+
					  "@bag to open your pokedex\n"+
					  "@catch to getx10 pokemon\n"+
					  "@list to list the players\n"+
					  "@invite +`username` to join the battle\n"+
					  "@quit to quit the game\n"+
					  "@search to search pokemon from pokedex\n"+
					  "Enter your choice: ")
			break
		}
	}

	go receiveMessages(conn)

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Error sending message: ", err)
			return
		} 
		if strings.Split(text, " ")[0] == "@battle"{
			break
		}
	}

	
}

func receiveMessages(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving message:", err)
			return
		}
		fmt.Println(string(buffer[:n]))
		if string(buffer[:n]) == "You out the game"{
			os.Exit(0)
			break
		}
	}
}
