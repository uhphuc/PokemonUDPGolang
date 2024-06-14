package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Pokedex struct {
	Id       string   `json:"ID"`
	Name     string   `json:"Name"`
	Types    []string `json:"types"`
	Link     string   `json:"URL"`
	PokeInfo PokeInfo `json:"Poke-Information"`
}
type PokeInfo struct {
	Hp          int         `json:"HP"`
	Atk         int         `json:"ATK"`
	Def         int         `json:"DEF"`
	SpAtk       int         `json:"Sp.Atk"`
	SpDef       int         `json:"Sp.Def"`
	Speed       int         `json:"Speed"`
	TypeDefense TypeDef     `json:"Type-Defenses"`
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

type ExpStats struct {
	GiveExp int
	GiveHP int 
	GiveATK int 
	GiveDef int
	GiveSpATK int
	GiveSpDef int 
	GiveSpeed int
}

func main() {

	fmt.Println("Connecting to the web")

	resp, err := http.Get("https://pokemondb.net/pokedex/national")
	if err != nil {
		fmt.Println("Error fetching Poke Homepage: ", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body: ", err)
		return
	}
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error parsing HTML: ", err)
		return
	}
	fmt.Println("Downloading... ")
	pokemons := getPokedex(doc)
	var allPoke []Pokedex
	var allPokeInfo PokeInfo

	for _, poke := range pokemons {

		allPokeInfo = getDetail(poke.Link)
		poke.PokeInfo = allPokeInfo
		allPoke = append(allPoke, poke)
		allPokeInfo = PokeInfo{}
	}

	jsonData, err := json.MarshalIndent(allPoke, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON: ", err)
		return
	}

	err = ioutil.WriteFile("pokedex.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file: ", err)
		return
	}

	fmt.Println("Pokedex data has been written to pokedex.json")
}

func getPokedex(n *html.Node) []Pokedex {
	var pokemon []Pokedex
	var currentPoke Pokedex
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "infocard-lg-data text-muted" {
					PokeId := strings.Split(getOnce(n, "small"), "a")[0]
					PokeId = strings.TrimSpace(PokeId)
					currentPoke.Id = PokeId
					pokeName := getInsideTag(n, "a", "class", "ent-name")
					pokeName = strings.TrimSpace(pokeName)
					currentPoke.Name = pokeName
					links := getStringElement(n, "a", "href")
					currentPoke.Link = strings.Split(links, "/type")[0]
					types := getInsideTag(n, "a", "class", "itype")
					typeSplit := strings.Split(types, " ")
					for _, eachType := range typeSplit {
						if eachType != "" {
							currentPoke.Types = append(currentPoke.Types, eachType)
						}
					}
					if currentPoke.Name != ""{
						pokemon = append(pokemon, currentPoke)
						currentPoke = Pokedex{}
					}

				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return pokemon
}

func getDetail(url string) PokeInfo {
	resp, err := http.Get("https://pokemondb.net" + url)
	if err != nil {
		fmt.Println("Error fetching WEBTOON homepage: ", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body: ", err)
		os.Exit(1)
	}
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error parsing HTML: ", err)
		os.Exit(1)
	}
	var pokeInfo PokeInfo
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "vitals-table" {
					result := getOnce(n, "th")
					stats := strings.Split(result, " ")
					for _, st := range stats {
						if st == "HP" {
							listStat := getStatNumber(n, "td", "class", "cell-num")
							for i := 0; i < len(listStat); i++ {
								pokeInfo.Hp = listStat[0]
								pokeInfo.Atk = listStat[3]
								pokeInfo.Def = listStat[6]
								pokeInfo.SpAtk = listStat[9]
								pokeInfo.SpDef = listStat[12]
								pokeInfo.Speed = listStat[15]
							}
						}
					}
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "table" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "type-table type-table-pokedex" {
					result := getOnce(n, "a")
					stats := strings.Split(result, " ")
					for _, st := range stats {
						if st == "Nor" {
							listDefRate := getRatioDef(n)
							for i := 0; i < len(listDefRate); i++ {
								pokeInfo.TypeDefense.Normal = float32(listDefRate[0]) / 100
								pokeInfo.TypeDefense.Fire = float32(listDefRate[1]) / 100
								pokeInfo.TypeDefense.Water = float32(listDefRate[2]) / 100
								pokeInfo.TypeDefense.Electric = float32(listDefRate[3]) / 100
								pokeInfo.TypeDefense.Grass = float32(listDefRate[4]) / 100
								pokeInfo.TypeDefense.Ice = float32(listDefRate[5]) / 100
								pokeInfo.TypeDefense.Fighting = float32(listDefRate[6]) / 100
								pokeInfo.TypeDefense.Poison = float32(listDefRate[7]) / 100
								pokeInfo.TypeDefense.Ground = float32(listDefRate[8]) / 100
							}
						}
						if st == "Fly" {
							listDefRate := getRatioDef(n)
							for i := 0; i < len(listDefRate); i++ {
								pokeInfo.TypeDefense.Flying = float32(listDefRate[0]) / 100
								pokeInfo.TypeDefense.Psychic = float32(listDefRate[1]) / 100
								pokeInfo.TypeDefense.Bug = float32(listDefRate[2]) / 100
								pokeInfo.TypeDefense.Rock = float32(listDefRate[3]) / 100
								pokeInfo.TypeDefense.Ghost = float32(listDefRate[4]) / 100
								pokeInfo.TypeDefense.Dragon = float32(listDefRate[5]) / 100
								pokeInfo.TypeDefense.Dark = float32(listDefRate[6]) / 100
								pokeInfo.TypeDefense.Steel = float32(listDefRate[7]) / 100
								pokeInfo.TypeDefense.Fairy = float32(listDefRate[8]) / 100
							}
						}
					}
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "tab-moves-21" {
					result := getOnce(n, "h3")
					contentResult := strings.Split(result, " ")
					for _, text := range contentResult {
						if text == "Moves" {

						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return pokeInfo
}

func getInsideTag(n *html.Node, data, key, val string) string {
	var result = ""
	if n.Type == html.ElementNode && n.Data == data {
		for _, attr := range n.Attr {
			if attr.Key == key && strings.Split(attr.Val, " ")[0] == val {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					result += " " + strings.TrimSpace(c.Data)
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += getInsideTag(c, data, key, val)
	}
	return result
}
func getOnce(n *html.Node, tagName string) string {
	var result string
	if n.Type == html.ElementNode && n.Data == tagName {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			result += " " + strings.TrimSpace(c.Data)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += getOnce(c, tagName)
	}
	return result
}
func getStringElement(n *html.Node, data, key string) string {
	var result string
	if n.Type == html.ElementNode && n.Data == data {
		for _, attr := range n.Attr {
			if attr.Key == key {
				result = attr.Val
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += getStringElement(c, data, key)
	}
	return result
}
func getStatNumber(n *html.Node, tagName, key, val string) []int {
	var numbers []int
	if n.Type == html.ElementNode && n.Data == tagName {
		for _, attr := range n.Attr {
			if attr.Key == key && attr.Val == val {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					number, _ := strconv.Atoi(strings.TrimSpace(c.Data))
					numbers = append(numbers, number)
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		number := getStatNumber(c, tagName, key, val)
		for _, c := range number {
			numbers = append(numbers, c)
		}
	}
	return numbers
}
func getRatioDef(n *html.Node) []int {
	var percentg []int
	if n.Type == html.ElementNode && n.Data == "td" {
		for _, attr := range n.Attr {
			if attr.Key == "class" {
				percent := strings.Split(attr.Val, "-")[4]
				number, _ := strconv.Atoi(percent)
				percentg = append(percentg, number)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		number := getRatioDef(c)
		for _, c := range number {
			percentg = append(percentg, c)
		}
	}
	return percentg
}
