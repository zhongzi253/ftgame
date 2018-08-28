package main

import (
	"./goals"
	"./utils"
	"./winlose"
	//"container/list"
	//"flag"
	//"fmt"
	//"github.com/PuerkitoBio/goquery"
	"log"
	//"net/http"
	//"net/smtp"
	"strconv"
	//"strings"
	"time"
)

func CoreLoop() {
	weakup := 0
	utils.SleepSleep(&weakup)
	for {
		RunOnce()
		utils.SleepSleep(&weakup)
	}
}

type GameInstance_t interface {
	RunOnce()
	TryRun()
	TestLoop()
}

var g_games [utils.TYPE_GAME_NUM]GameInstance_t

func FactoryGame(type_of_game int) GameInstance_t {
	switch type_of_game {
	case utils.TYPE_GAME_GOALS:
		return goals.NewGame()
	case utils.TYPE_GAME_WINLOSE:
		return winlose.NewGame()
	default:
		panic("Wrong game type:" + strconv.Itoa(type_of_game))
	}
}

func RunFactory(type_of_game int) {
	if type_of_game == -1 {
		for i := 0; i < utils.TYPE_GAME_NUM; i++ {
			g_games[i] = FactoryGame(i)
		}
	} else {
		g_games[type_of_game] = FactoryGame(type_of_game)
	}

}
func RunOnce() {

	old := time.Now()
	log.Printf("Start to Run once, %v\n", time.Unix(old.Unix(), 0))
	for _, game := range g_games {
		if game != nil {
			game.RunOnce()
		}
	}
	log.Printf("DONE for this RunOnce, cost %v\n\n", time.Now().Sub(old))
}

func TryRun() {
	a := 0
	utils.SleepSleep(&a)
	log.Println(a)
	for _, game := range g_games {
		if game != nil {
			game.TryRun()
		}
	}
}

func TestLoop() {
	for _, game := range g_games {
		if game != nil {
			game.TestLoop()
		}
	}
}
func init() {

	utils.ParseFlag()

}
func main() {
	RunFactory(utils.FlagGameType)
	switch utils.FlagMode {
	case 1:
		RunOnce()
	case 2:
		CoreLoop()
	case 3:
		TestLoop()
	case 4:
		TryRun()
	}
}
