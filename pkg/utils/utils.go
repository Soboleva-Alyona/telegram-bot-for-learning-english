package utils

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
	"time"
)

const (
	AddCommand = "add"

	DeleteCommand = "delete"

	StartCommand = "start"

	DropBaseCommand = "drop"

	InfoCommand = "info"

	WrongFormat = "String doesn't meet the format"

	SuccessText = "success"

	SomeError = "some error, try again"

	InfoTextPath = "pkg/utils/info.txt"

	ConfirmCommand = "confirm"

	AskConfirmDropBaseMessage = "To confirm drop of the whole base type \"/confirm drop\" "
)

var ConfirmationKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Confirm", "Confirm"),
		tgbotapi.NewInlineKeyboardButtonData("Cancel", "Cancel"),
	),
)

type WordAndTranslation struct {
	Word        string
	Translation string
}

func FormatDateForDB(t time.Time) string {
	return fmt.Sprintf("%d-%d-%d %d:%d:%d\n",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Hour(),
		t.Second())
}

func ParseStringToWordAndTranslation(str string) (successText string, word string, translation string) {
	splitStr := strings.Split(str, " - ")

	if len(splitStr) == 2 {
		return SuccessText, splitStr[0], splitStr[1]
	} else {
		return WrongFormat, "", ""
	}
}
