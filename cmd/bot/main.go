package main

import (
	"bufio"
	_ "github.com/go-telegram-bot-api/telegram-bot-api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"zetellBot/pkg/database"
	"zetellBot/pkg/telegram"
)

func main() {
	botToken := ""
	f, e := os.Open("./cmd/bot/token.txt")
	if e != nil {
		log.Fatal("Can't open file", e.Error())
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		botToken = scanner.Text()
	}

	bot, err := tgbotapi.NewBotAPI(botToken)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	err = database.ConnectAndGet()

	if err != nil {
		log.Panic("Can't connect to db" + err.Error())
	}

	telegramBot := telegram.NewBot(bot)
	if err := telegramBot.Start(); err != nil {
		log.Fatal(err)
	}

}
