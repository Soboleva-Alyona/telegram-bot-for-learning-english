package telegram

import (
	_ "database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
	"time"
	"zetellBot/pkg/database"
	"zetellBot/pkg/utils"
)

type Bot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(bot *tgbotapi.BotAPI) *Bot {
	return &Bot{bot: bot}
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)
	updates, err := b.initUpdatesChan()
	if err != nil {
		return err
	}
	b.handleUpdates(updates)
	return nil

}

func (b *Bot) initUpdatesChan() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, e := b.bot.GetUpdatesChan(u)
	if e != nil {
		return nil, e
	}
	return updates, nil
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
			continue
		}
		b.sendMessage(update.Message, "Type a command, please.\nType /info for help.")
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	userId := message.From.ID
	userIdStr := strconv.Itoa(userId)
	command := message.Command()
	var text string
	switch command {
	case utils.InfoCommand:
		text = getInfoText(text)
	case utils.StartCommand:
		text, _ = b.initUserTableOrRefuse(userIdStr)
		go b.sendWordsWithIntervals(message, userIdStr)
	case utils.AddCommand, utils.DeleteCommand, utils.DropBaseCommand:
		var exists bool
		text, exists = b.checkIfUserAuthorizedAndSetAResponseIfNot(userIdStr)
		if !exists {
			text = "You haven't made your base"
			break
		}
		switch command {
		case utils.AddCommand:
			if len(message.Text) != 0 {
				// TODO("confirmation")
				//msg := tgbotapi.NewMessage(message.Chat.ID, text)
				//msg.ReplyMarkup = utils.ConfirmationKeyboard
				successOrNot, err := b.addWordAndTranslation(userIdStr, message.CommandArguments())

				if err != nil {
					text = utils.SomeError
				}
				if successOrNot == utils.SuccessText {
					text = "The word was added"
				} else {
					text = successOrNot
				}
			} else {
				text = "You haven't entered a word to add"
			}

		case utils.DropBaseCommand:
			err := database.DropTableForUser(userIdStr)
			if err != nil {
				text = "Problem while drop occur, try again later"
			} else {
				text = "Your word base was successfully dropped"
			}
		case utils.DeleteCommand:
			if len(message.Text) != 0 {
				successOrNot, err := b.deleteWordAndTranslation(userIdStr, message.CommandArguments())
				if err != nil {
					text = utils.SomeError
				}
				if successOrNot == utils.SuccessText {
					text = "The word was deleted"
				} else {
					text = successOrNot
				}
			} else {
				text = "You haven't entered a word and translation to delete"
			}
		}
		break
	default:
		text = "No such command."
	}
	b.sendMessage(message, text)
}

func getInfoText(text string) string {
	bytes, _ := os.ReadFile(utils.InfoTextPath)
	text = string(bytes)
	return text
}

func (b *Bot) checkIfUserAuthorizedAndSetAResponseIfNot(userIdStr string) (string, bool) {
	var text string
	exists := b.checkUserTable(userIdStr)
	if !exists {
		text = "You are not authorized."
	}
	return text, exists
}

func (b *Bot) deleteWordAndTranslation(userIdStr string, wordAndTranslation string) (string, error) {
	successOrNotText, err := database.DeleteWordAndTranslation(userIdStr, wordAndTranslation)

	if err != nil {
		return utils.SomeError, err
	}
	return successOrNotText, nil
}

func (b *Bot) addWordAndTranslation(userIdStr string, wordAndTranslation string) (string, error) {
	successOrNotText, err := database.AddWordAndTranslation(userIdStr, wordAndTranslation)

	if err != nil {
		return utils.SomeError, err
	}
	return successOrNotText, nil
}

func (b *Bot) initUserTableOrRefuse(userIdStr string) (string, bool) {
	res := b.checkUserTable(userIdStr)

	if res {
		return "You already have your base.", true
	} else {
		err := database.CreateTableForUser(userIdStr)
		if err != nil {
			log.Fatal("Error while creating table " + err.Error())
		}
		return "Your new base successfully created.", false
	}

}

func (b *Bot) checkUserTable(userIdStr string) bool {
	res, err := database.CheckIfTableExists("user_" + userIdStr)
	if err != nil {
		log.Fatal("Error while checking user's table for exist")
	}
	return res
}

func (b *Bot) sendMessage(message *tgbotapi.Message, text string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "HTML"
	msg.ReplyToMessageID = message.MessageID
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Bot) sendWordsWithIntervals(message *tgbotapi.Message, userIdStr string) {
	for {
		exists := b.checkUserTable(userIdStr)
		if !exists {
			break
		}
		wordToSend := selectWordToSend(userIdStr)
		if len(wordToSend.Word) == 0 {
			time.Sleep(time.Minute * 20) // minimal sending interval, before that we would not need to send something
			continue
		}
		b.sendMessage(message, wordToSend.Word+` <tg-spoiler>`+wordToSend.Translation+`</tg-spoiler>`)
	}
}

func selectWordToSend(userIdStr string) utils.WordAndTranslation {
	var wordToSend utils.WordAndTranslation
	curTime := time.Now()

	wordToSend = database.SelectWordForTime(userIdStr, curTime)
	return wordToSend
}
