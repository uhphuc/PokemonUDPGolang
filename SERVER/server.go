package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

)

type Client struct {
	Name            string
	Addr            *net.UDPAddr
	userCurrentPoke Pokedex
	userPokedex     []Pokedex
	currentPoke     Pokedex
	battlePoke      []Pokedex
}
type Battle struct {
	Player1      *Client
	Player2      *Client
	CurrentPoke1 *Pokedex
	CurrentPoke2 *Pokedex
	CurrentTurn  *net.UDPAddr
}
type Pokedex struct {
	Id       string   `json:"ID"`
	Name     string   `json:"Name"`
	Level    int      `json:"Level"`
	Exp		int
	Types    []string `json:"types"`
	Link     string   `json:"URL"`
	PokeInfo PokeInfo `json:"Poke-Information"`
}
type PokeInfo struct {
	Hp          int     `json:"HP"`
	Atk         int     `json:"ATK"`
	Def         int     `json:"DEF"`
	SpAtk       int     `json:"Sp.Atk"`
	SpDef       int     `json:"Sp.Def"`
	Speed       int     `json:"Speed"`
	TypeDefense TypeDef `json:"Type-Defenses"`
}
type TypeDef struct {
	Normal   float32
	Fire     float32
	Water    float32
	Electric float32
	Grass    float32
	Ice      float32
	Fighting float32
	Poison   float32
	Ground   float32
	Flying   float32
	Psychic  float32
	Bug      float32
	Rock     float32
	Ghost    float32
	Dragon   float32
	Dark     float32
	Steel    float32
	Fairy    float32
}

var (
	clients    = make(map[string]*Client)
	pokedex    []Pokedex
	battles    = make(map[string]*Client)
	mu         sync.Mutex
	invitation = make(map[string]string)
	games      = make(map[string]*Battle)
	state      *net.UDPAddr
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Server is running on port 8080")

	buffer := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue
		}
		message := string(buffer[:n])
		handleMessage(message, addr, conn)
	}
}

func handleMessage(message string, addr *net.UDPAddr, conn *net.UDPConn) {
	mu.Lock()
	defer mu.Unlock()

	if strings.HasPrefix(message, "@") {
		parts := strings.Split(message, " ")
		command := parts[0]
		senderName := getUsernameByAddr(addr)

		client := clients[senderName]

		switch command {
		case "@join":
			if checkExist(parts[1]) {
				sendMessageToClient("Invalid", addr, conn)
			} else {
				username := parts[1]
				clients[username] = &Client{Name: username, Addr: addr}
				fmt.Printf("User [%s] joined\n", username)
				sendMessageToClient("["+username+"] Welcome to game POKE BATTLE", addr, conn)
			}
		case "@quit":
			username := getUsernameByAddr(addr)
			delete(clients, username)
			fmt.Print("User [" + username + "] out the game\n")
			sendMessageToClient("You are out the game", addr, conn)
		case "@choose":
			if parts[1] == "1" {
				OpenFile("Database\\pokedex.json", &pokedex)
				for _, poke := range pokedex {
					if poke.Id == "#0001" {
						client.userCurrentPoke = poke
						client.userCurrentPoke.Level = 1
						client.userPokedex = append(client.userPokedex, client.userCurrentPoke)
					}
				}
				sendMessageToClient("Valid", addr, conn)
			} else if parts[1] == "2" {
				OpenFile("Database\\pokedex.json", &pokedex)
				for _, poke := range pokedex {
					if poke.Id == "#0004" {
						client.userCurrentPoke = poke
						client.userCurrentPoke.Level = 1
						client.userPokedex = append(client.userPokedex, client.userCurrentPoke)
					}
				}
				sendMessageToClient("Valid", addr, conn)
			} else if parts[1] == "3" {
				OpenFile("Database\\pokedex.json", &pokedex)
				for _, poke := range pokedex {
					if poke.Id == "#0007" {
						client.userCurrentPoke = poke
						client.userCurrentPoke.Level = 1
						client.userPokedex = append(client.userPokedex, client.userCurrentPoke)
					}
				}
				sendMessageToClient("Valid", addr, conn)
			} else {
				sendMessageToClient("Cannot", addr, conn)
			}
			CreateFile(senderName+"_Pokedex.json", client.userPokedex)
		case "@search": 
			if len(parts) != 2 {
				sendMessageToClient("Invalid command! Please try again!\n", addr, conn)
			} else {
				OpenFile("Database\\pokedex.json", &pokedex)
				for _, poke := range pokedex {
					if poke.Name == parts[1] {
						msg := fmt.Sprintf("ID: %s - Name: %s - HP: %d - ATK: %d - DEF: %d - SPEED: %d\n",
							poke.Id, poke.Name, poke.PokeInfo.Hp, poke.PokeInfo.Atk, poke.PokeInfo.Def, poke.PokeInfo.Speed)
					
					sendMessageToClient(msg, addr, conn)	
					}
				}
				
			}
		case "@catch":
			getPoke := Rollx10Poke(client.userCurrentPoke)
			ListPokemon := "Your new pokemon:\n"
			for _, poke := range getPoke {
				ListPokemon += fmt.Sprintf("[Name: %s --- Level: %d]\n", poke.Name, poke.Level)
			}
			sendMessageToClient(ListPokemon, addr, conn)
			client.userPokedex = append(client.userPokedex, getPoke...)
			CreateFile(senderName+"_Pokedex.json", client.userPokedex)
		case "@bag":
			msg := "Your Bag:\n"
			for _, poke := range client.userPokedex {
				msg += fmt.Sprintf("ID: %s - Name: %s [Level: %d] - HP: %d - ATK: %d - DEF: %d - SPEED: %d\n",
					poke.Id, poke.Name, poke.Level, poke.PokeInfo.Hp, poke.PokeInfo.Atk, poke.PokeInfo.Def, poke.PokeInfo.Speed)
			}
			sendMessageToClient(msg, addr, conn)
		case "@pick":
			if len(parts) != 4 {
				sendMessageToClient("Invalid input! Please try again!\n", addr, conn)
			} else {
				confirm := "Your pokemon choosen:\n"
				if checkPokeExist(parts[1], parts[2], parts[3], client) {
					for _, poke := range client.userPokedex {
						if parts[1] == poke.Id {
							confirm += poke.Name + " "
							client.battlePoke = append(client.battlePoke, poke)
						}
						if parts[2] == poke.Id {
							confirm += poke.Name + " "
							client.battlePoke = append(client.battlePoke, poke)
						}
						if parts[3] == poke.Id {
							confirm += poke.Name + " "
							client.battlePoke = append(client.battlePoke, poke)
						}
					}
					confirm += "\n(Usage: @play to start!)\n"
					sendMessageToClient(confirm, addr, conn)
				} else {
					sendMessageToClient("Poke you choose is not have in your pokedex!", addr, conn)
				}
			}
		case "@list":
			competitors := "Current player:\n"
			for _, user := range clients {
				if user.Name != senderName {
					competitors += fmt.Sprintf("[Player: %s]\n", user.Name)
				}
			}
			sendMessageToClient(competitors, addr, conn)
		case "@invite":
			for _, user := range clients {
				if user.Name == parts[1] || parts[1]!= senderName { // if exist username like this then
					for _, bt := range battles{
						if user == bt {
							sendMessageToClient(parts[1]+" is in battle, please try later!", addr, conn)
							return
						}
					}
					invitation[addr.String()] = senderName
					invitation[user.Addr.String()] = parts[1] // invitation with index string of that user addr ---> get value of part[1]
					
				} else if parts[1] == senderName {
					sendMessageToClient("Cannot invite yourself!!!", addr, conn)
					return
				}
			}
			sendMessageToClient("Waiting for your competitor!", addr, conn)
			sendPrivateMessage(senderName+" send you a request to battle!(@accept yes/no)\n",parts[1], conn, addr)
			// I think error happen when 2 clients @invite will add into battle
		case "@accept":
			if strings.ToLower(parts[1]) == "yes" {
				receiverName := invitation[addr.String()]
				var inviterName string
				for _, invite := range invitation {
					if receiverName != invite {
						inviterName = invite
					}
				}
				for _, user := range clients {
					if user.Name == inviterName {
						sendMessageToClient(senderName+" has accepted the battle\n(Usage: @pick #id_pokemon1 #id_pokemon2 #id_pokemon3)\n", user.Addr, conn)
						battles[user.Addr.String()] = user // inviter client

					}
					if user.Addr.String() == addr.String() {
						battles[addr.String()] = user // receiver client
					}
				}
				sendMessageToClient("You are join the battle!\n(Usage: @pick #id_pokemon1 #id_pokemon2 #id_pokemon3)\n", addr, conn)
			} else if strings.ToLower(parts[1]) == "no" {
				var inviterName string
				for _, invite := range invitation {
					if senderName != invite {
						inviterName = invite
					}
				}
				for _, user := range clients {
					if user.Name == inviterName {
						sendMessageToClient("Your competitor is decline\nChoose another user or other task", user.Addr, conn)
					}
				}
				delete(invitation, addr.String())
				sendMessageToClient("You decline successfull\nLet continue other tasks\n", addr, conn)
			} else {
				sendMessageToClient("Invalid command!\n", addr, conn)
			}
		case "@play":
			player, inBattle := battles[addr.String()]
			if !inBattle {
				sendMessageToClient("You are not in the battle! Cannot use this command!", addr, conn)
				return
			}
			gameKey := ""
			var opponent *Client
			for _, bat := range battles {
				if bat.Addr.String() != addr.String() {
					opponent = battles[bat.Addr.String()]
					gameKey = fmt.Sprintf("%s:%s", player.Name, opponent.Name)
				}
			}
			if _, inBattle := games[gameKey]; inBattle {
				sendMessageToClient("A game is already in process.", addr, conn)
				return
			}
			if opponent.battlePoke[0].PokeInfo.Speed > player.battlePoke[0].PokeInfo.Speed {
				state = opponent.Addr
			} else {
				state = player.Addr
			}
			sendMessageToClient("You first", state, conn)

			battle := &Battle{
				Player1:      player,
				Player2:      opponent,
				CurrentPoke1: &player.battlePoke[0],
				CurrentPoke2: &opponent.battlePoke[0],
				CurrentTurn:  state,
			}
			games[gameKey] = battle
		case "@attack":
			handleAttack(addr, conn)
		case "@switch":
			if len(parts) != 2 {
				sendMessageToClient("Invalid command!", addr, conn)
				return
			}
			handleSwitch(conn, addr, parts[1])
		case "@surrender":
			opponent, inBattle := battles[addr.String()]

			if !inBattle {
				sendMessageToClient("You are not in the battle! Cannot use this command!", addr, conn)
				return
			}
			var player *Client
			gameKey := ""
			for _, bat := range battles {
				if bat.Addr.String() != addr.String() {
					player = battles[bat.Addr.String()]
					gameKey = fmt.Sprintf("%s:%s", player.Name, opponent.Name)
				}
			}
			game, inGame := games[gameKey]
			if !inGame {
				sendMessageToClient("You are not already\n(Usage: @play to ready the battle)", addr, conn)
				return
			}
				
			winner := player
			loser := opponent
			conn.WriteToUDP([]byte(fmt.Sprintf("Game over! %s wins!", winner.Name)), winner.Addr)
			conn.WriteToUDP([]byte(fmt.Sprint("Game over! You lose!", loser.Name)), loser.Addr)

			
			delete(games, gameKey)
			delete(battles, game.Player1.Addr.String())
			delete(battles, game.Player2.Addr.String())
		default:
			sendMessageToClient("Invalid command", addr, conn)
		}
	} else {
		sendMessageToClient("Invalid command format", addr, conn)
	}
}

func getUsernameByAddr(addr *net.UDPAddr) string {
	for _, client := range clients {
		if client.Addr.IP.Equal(addr.IP) && client.Addr.Port == addr.Port {
			return client.Name
		}
	}
	return ""
}

func sendMessageToClient(message string, addr *net.UDPAddr, conn *net.UDPConn) {
	_, err := conn.WriteToUDP([]byte(message), addr)
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func checkExist(name string) bool {
	_, exist := clients[name]
	if exist {
		return true
	} else {
		return false
	}
}

func sendPrivateMessage(message, recipient string, conn *net.UDPConn, addr *net.UDPAddr) {
	client, exists := clients[recipient]
	if !exists {
		fmt.Println("Recipient not found:", recipient)
		conn.WriteToUDP([]byte("NotFound"), addr)
		return
	}
	conn.WriteToUDP([]byte(message), client.Addr)
}
func OpenFile(fileName string, key interface{}) {
	inFile, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Error to open file: ", err)
		os.Exit(1)
	}
	decoder := json.NewDecoder(inFile)
	err = decoder.Decode(key)
	if err != nil {
		log.Fatal("Error to open file: ", err)
		os.Exit(1)
	}
	inFile.Close()
}
func CreateFile(fileName string, key interface{}) {
	jsonData, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON: ", err)
		return
	}
	err = ioutil.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file: ", err)
		return
	}
	fmt.Println(fileName + " updated!")
}
func Rollx10Poke(userCurrentPoke Pokedex) []Pokedex {
	var userPokedex []Pokedex
	for i := 0; i < 10; i++ {
		getId := rand.Intn(1025-1) + 1
		var Idpoke string
		if getId < 10 {
			Idpoke = "#000" + strconv.Itoa(getId)
		} else if getId >= 10 && getId < 100 {
			Idpoke = "#00" + strconv.Itoa(getId)
		} else if getId >= 100 && getId < 1000 {
			Idpoke = "#0" + strconv.Itoa(getId)
		} else {
			Idpoke = "#" + strconv.Itoa(getId)
		}
		for _, poke := range pokedex {
			if Idpoke == poke.Id {
				userCurrentPoke = poke
				userCurrentPoke.Level = 1
				userPokedex = append(userPokedex, userCurrentPoke)
			}
		}
	}
	return userPokedex
}
func checkPokeExist(poke1, poke2, poke3 string, client *Client) bool {
	var allExist = false
	epoke1, epoke2, epoke3 := false, false, false
	for _, poke := range client.userPokedex {
		if poke1 == poke.Id {
			epoke1 = true
		}
		if poke2 == poke.Id {
			epoke2 = true
		}
		if poke3 == poke.Id {
			epoke3 = true
		}
	}
	if epoke1 && epoke2 && epoke3 {
		allExist = true
	}
	return allExist
}
func handleAttack(addr *net.UDPAddr, conn *net.UDPConn) {

	opponent, inBattle := battles[addr.String()]
	if !inBattle {
		sendMessageToClient("You are not in the battle! Cannot use this command!", addr, conn)
		return
	}
	gameKey := ""
	var player *Client
	for _, bat := range battles {
		if bat.Addr.String() != addr.String() {
			player = battles[bat.Addr.String()]
			gameKey = fmt.Sprintf("%s:%s", player.Name, opponent.Name)
		}
	}
	game, inGame := games[gameKey]
	if !inGame {
		sendMessageToClient("You are not already\n(Usage: @play to ready the battle)", addr, conn)
		return
	}

	if addr.String() != state.String() {
		sendMessageToClient("Not your turn!", addr, conn)
		return
	}

	if addr.String() == game.Player1.Addr.String() {
		check := rand.Intn(2- 0)+0
		if check == 0 {
			damageNormal, _ := getDmgNumber(*game.CurrentPoke1, *game.CurrentPoke2)
			game.CurrentPoke2.PokeInfo.Hp -= damageNormal
			state = game.Player2.Addr
		} else {
			_, specialDmg := getDmgNumber(*game.CurrentPoke1, *game.CurrentPoke2)
			game.CurrentPoke2.PokeInfo.Hp -= specialDmg
			state = game.Player2.Addr
		}
	} else if addr.String() == game.Player2.Addr.String() {
		check := rand.Intn(2-0) + 0
		if check == 0 {
			damageNormal, _ := getDmgNumber(*game.CurrentPoke2, *game.CurrentPoke1)
			game.CurrentPoke1.PokeInfo.Hp -= damageNormal
			state = game.Player1.Addr
		} else {
			_, specialDmg := getDmgNumber(*game.CurrentPoke2, *game.CurrentPoke1)
			game.CurrentPoke1.PokeInfo.Hp -= specialDmg
			state = game.Player1.Addr
		}
	}

	if game.CurrentPoke1.PokeInfo.Hp <= 0 {
		var temp []Pokedex
		if len(game.Player1.battlePoke) > 1{
			temp = append(temp, game.Player1.battlePoke[1:]...)
			game.CurrentPoke1 = &temp[0]
		}
		game.Player1.battlePoke = temp
	} else if game.CurrentPoke2.PokeInfo.Hp <= 0 {
		var temp []Pokedex
		if len(game.Player2.battlePoke) > 1{
			temp = append(temp, game.Player2.battlePoke[1:]...)
			game.CurrentPoke2 = &temp[0]
		}
		game.Player2.battlePoke = temp
	}

	if game.Player1.battlePoke == nil {
		winner := game.Player2
		loser := game.Player1
		conn.WriteToUDP([]byte(fmt.Sprintf("Game over! %s wins!", winner.Name)), winner.Addr)
		conn.WriteToUDP([]byte(fmt.Sprint("Game over! You lose!", loser.Name)), loser.Addr)

		delete(games, gameKey)
		delete(battles, game.Player1.Addr.String())
		delete(battles, game.Player2.Addr.String())
	} else if game.Player2.battlePoke == nil {
		winner := game.Player1
		loser := game.Player2
		conn.WriteToUDP([]byte(fmt.Sprintf("Game over! %s wins!", winner.Name)), winner.Addr)
		conn.WriteToUDP([]byte(fmt.Sprint("Game over! You lose!", loser.Name)), loser.Addr)

		delete(games, gameKey)
		delete(battles, game.Player1.Addr.String())
		delete(battles, game.Player2.Addr.String())
	} else {
		if game.CurrentPoke2.PokeInfo.Hp < 0{
			game.CurrentPoke2.PokeInfo.Hp = 0
		}
		if game.CurrentPoke1.PokeInfo.Hp < 0{
			game.CurrentPoke1.PokeInfo.Hp = 0
		}
		conn.WriteToUDP([]byte(fmt.Sprintf("%s attacked %s!Your %s's HP: %d\n%s's opponent - HP: %d",
			game.CurrentPoke2.Name, game.CurrentPoke1.Name,
			game.CurrentPoke2.Name, game.CurrentPoke2.PokeInfo.Hp,
			game.CurrentPoke1.Name, game.CurrentPoke1.PokeInfo.Hp)), opponent.Addr)
		conn.WriteToUDP([]byte(fmt.Sprintf("%s attacked and remaining HP: %d", game.CurrentPoke1.Name, game.CurrentPoke1.PokeInfo.Hp)), player.Addr)
	}
}
func handleSwitch(conn *net.UDPConn, addr *net.UDPAddr, id string) {
	opponent, inBattle := battles[addr.String()]
	if !inBattle {
		sendMessageToClient("You are not in the battle! Cannot use this command!", addr, conn)
		return
	}
	gameKey := ""
	var player *Client
	for _, bat := range battles {
		if bat.Addr.String() != addr.String() {
			player = battles[bat.Addr.String()]
			gameKey = fmt.Sprintf("%s:%s", player.Name, opponent.Name)
		}
	}
	game, inGame := games[gameKey]
	if !inGame {
		sendMessageToClient("You are not already\n(Usage: @play to ready the battle)", addr, conn)
		return
	}
	if addr.String() != state.String() {
		sendMessageToClient("Not your turn!", addr, conn)
		return
	}
	if addr.String() == game.Player1.Addr.String() {
		
		for _, poke := range game.Player1.battlePoke {
			if poke.Id == id {
				game.CurrentPoke1 = &poke
				sendMessageToClient(fmt.Sprintf("You switch successfully\n<%s's> in the battle", game.CurrentPoke1.Name), game.Player1.Addr, conn)
				sendMessageToClient(fmt.Sprintf("Your competitor switch --> <%s> in the battle", game.CurrentPoke1.Name), game.Player2.Addr, conn)
				state = game.Player2.Addr
			}
		}
	} else if addr.String() == game.Player2.Addr.String() {
		for _, poke := range game.Player2.battlePoke {
			if poke.Id == id {
				game.CurrentPoke2 = &poke
				sendMessageToClient(fmt.Sprintf("You switch successfully\n<%s's> in the battle", game.CurrentPoke2.Name), game.Player2.Addr, conn)
				sendMessageToClient(fmt.Sprintf("Your competitor switch --> <%s> in the battle", game.CurrentPoke2.Name), game.Player1.Addr, conn)
				state = game.Player1.Addr
			}
		}
	}
}

func getDmgNumber(pAtk Pokedex, pRecive Pokedex) (int, int) {
	var types = make(map[string]float32)

	types["Normal"] = pRecive.PokeInfo.TypeDefense.Normal
	types["Fire"] = pRecive.PokeInfo.TypeDefense.Fire
	types["Water"] = pRecive.PokeInfo.TypeDefense.Water
	types["Electric"] = pRecive.PokeInfo.TypeDefense.Electric
	types["Grass"] = pRecive.PokeInfo.TypeDefense.Grass
	types["Ice"] = pRecive.PokeInfo.TypeDefense.Ice
	types["Fighting"] = pRecive.PokeInfo.TypeDefense.Fighting
	types["Poison"] = pRecive.PokeInfo.TypeDefense.Poison
	types["Ground"] = pRecive.PokeInfo.TypeDefense.Ground
	types["Flying"] = pRecive.PokeInfo.TypeDefense.Flying
	types["Psychic"] = pRecive.PokeInfo.TypeDefense.Psychic
	types["Bug"] = pRecive.PokeInfo.TypeDefense.Bug
	types["Rock"] = pRecive.PokeInfo.TypeDefense.Rock
	types["Ghost"] = pRecive.PokeInfo.TypeDefense.Ghost
	types["Dragon"] = pRecive.PokeInfo.TypeDefense.Dragon
	types["Dark"] = pRecive.PokeInfo.TypeDefense.Dark
	types["Steel"] = pRecive.PokeInfo.TypeDefense.Steel
	types["Fairy"] = pRecive.PokeInfo.TypeDefense.Fairy

	var normal float32
	var special float32

	normal = float32(pAtk.PokeInfo.Atk) - float32(pRecive.PokeInfo.Def)

	var typeDefense float32 = 0.0
	for _, pAtkTypes := range pAtk.Types {
		for typeDef, def := range types {
			if typeDef == pAtkTypes {
				if typeDefense < def {
					typeDefense = def
				}
			}
		}
	}
	if normal < 0 || special < 0 {
		return 0, 0
	}
	special = float32(pAtk.PokeInfo.SpAtk)*typeDefense - float32(pRecive.PokeInfo.SpDef)
	return int(normal), int(special)
}
func getLevelExp(level int) (int, int) {

	totalExp := (level + 1) * (level + 1) * (level + 1)
	ExpAtLevel := totalExp - (level * level * level)
	if level == 1 {
		ExpAtLevel = totalExp
	}
	return totalExp, ExpAtLevel
}
