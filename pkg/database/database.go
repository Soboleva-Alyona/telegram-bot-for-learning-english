package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"zetellBot/pkg/utils"
	_ "zetellBot/pkg/utils"
)

const (
	dbName   = "postgres"
	password = "postgres"
	user     = "esoboleva"
	host     = "127.0.0.1"
	port     = "5432"
)

var db *sql.DB

func ConnectAndGet() (e error) {
	database, e := sql.Open("postgres",
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host,
			port,
			user,
			password,
			dbName))
	if e != nil {
		return e
	}
	db = database
	return nil
}

func CheckIfTableExists(name string) (bool, error) {
	res, err := db.Query(`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = $1);`, name)
	if err != nil {
		log.Fatal(err)
		return false, err

	}
	var answer bool
	for res.Next() {
		err = res.Scan(&answer)
		if err != nil {
			return false, err
		}
	}

	return answer, nil
}

func CreateTableForUser(id string) error {
	_, err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		"id" serial NOT NULL,
		"word" varchar(255),
		"translation" TEXT,
		"addition_time" TIMESTAMP,
    	"yet_need_to_send_min20" BOOLEAN,
    	"yet_need_to_send_hour" BOOLEAN,
        "yet_need_to_send_hour8" BOOLEAN,
        "yet_need_to_send_day" BOOLEAN,
        "yet_need_to_send_day3" BOOLEAN,
		"yet_need_to_send" BOOLEAN,
		PRIMARY KEY ("id")
	) WITH (
		OIDS=FALSE
	);`, "user_"+id))

	return err
}

func DropTableForUser(id string) error {
	_, err := db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s `, "user_"+id))

	return err
}

func DeleteWordAndTranslation(id string, wordAndTranslation string) (string, error) {
	text, word, translation := utils.ParseStringToWordAndTranslation(wordAndTranslation)

	if text == utils.WrongFormat {
		return text, nil
	}

	_, err := db.Exec(
		fmt.Sprintf(`
		DELETE FROM %s 
		WHERE word='%s' AND translation='%s' AND yet_need_to_send=TRUE
		`,
			"user_"+id, word, translation))
	if err != nil {
		return "", err
	}
	return utils.SuccessText, err
}

func AddWordAndTranslation(id string, wordAndTranslation string) (string, error) {
	now := time.Now()
	text, word, translation := utils.ParseStringToWordAndTranslation(wordAndTranslation)

	if text == utils.WrongFormat {
		return text, nil
	}

	_, err := db.Exec(fmt.Sprintf(`
	INSERT INTO %s (
					word, 
	                translation, 
	                addition_time, 
	                yet_need_to_send_min20, 
	                yet_need_to_send_hour,
	                yet_need_to_send_hour8,
	                yet_need_to_send_day,
	                yet_need_to_send_day3,
	                yet_need_to_send)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`, "user_"+id), word, translation, now, true, true, true, true, true, true)

	if err != nil {
		return utils.SomeError, err
	}

	return utils.SuccessText, nil
}

func SelectWordForTime(userIdStr string, curTime time.Time) utils.WordAndTranslation {
	rows, e := db.Query(fmt.Sprintf(`SELECT 
		word, 
		translation, 
		addition_time,
		yet_need_to_send_min20, 
		yet_need_to_send_hour,
		yet_need_to_send_hour8,
		yet_need_to_send_day,
		yet_need_to_send_day3 
		FROM %s WHERE yet_need_to_send = TRUE`, "user_"+userIdStr))

	if e != nil {
		log.Fatal("Problems while selecting words from DB: " + e.Error())
	}

	var wordToSend utils.WordAndTranslation

	for rows.Next() {
		var word string
		var translation string
		var additionTime time.Time
		var yetNeedToSendMin20,
			yetNeedToSendHour,
			yetNeedToSendHour8,
			yetNeedToSendDay,
			yetNeedToSendDay3 bool

		err := rows.Scan(
			&word,
			&translation,
			&additionTime,
			&yetNeedToSendMin20,
			&yetNeedToSendHour,
			&yetNeedToSendHour8,
			&yetNeedToSendDay,
			&yetNeedToSendDay3)

		if err != nil {
			log.Panic("Problems while scanning values from DB query")
			return utils.WordAndTranslation{}
		}

		Time20Min := additionTime.Add(time.Minute * 20)
		TimeHour := Time20Min.Add(time.Hour)
		Time8Hours := TimeHour.Add(time.Hour * 8)
		TimeDay := Time8Hours.Add(time.Hour * 24)
		Time3Day := TimeDay.Add(time.Hour * 24 * 3)
		TimeWeek := Time3Day.Add(time.Hour * 24 * 7)

		curTimeStr := curTime.Format(time.DateTime)
		curTime, err = time.Parse(time.DateTime, curTimeStr)

		if err != nil {
			log.Panic("Error while formatting date")
		}

		if curTime.After(TimeWeek) {
			wordToSend = getWordAndUpdateYetNeedToSend(userIdStr, word, translation, wordToSend, "yet_need_to_send")
			deleteWordAndTranslationAfterLastSending(userIdStr, word, translation)
			break
		}
		if curTime.After(Time3Day) && yetNeedToSendDay3 {
			wordToSend = getWordAndUpdateYetNeedToSend(userIdStr, word, translation, wordToSend, "yet_need_to_send_day3")
			break
		}
		if curTime.After(TimeDay) && yetNeedToSendDay {
			wordToSend = getWordAndUpdateYetNeedToSend(userIdStr, word, translation, wordToSend, "yet_need_to_send_day")
			break
		}
		if curTime.After(Time8Hours) && yetNeedToSendHour8 {
			wordToSend = getWordAndUpdateYetNeedToSend(userIdStr, word, translation, wordToSend, "yet_need_to_send_hour8")
			break
		}
		if curTime.After(TimeHour) && yetNeedToSendHour {
			wordToSend = getWordAndUpdateYetNeedToSend(userIdStr, word, translation, wordToSend, "yet_need_to_send_hour")
			break
		}
		if curTime.After(Time20Min) && yetNeedToSendMin20 {
			wordToSend = getWordAndUpdateYetNeedToSend(userIdStr, word, translation, wordToSend, "yet_need_to_send_min20")
			break
		}
	}
	return wordToSend
}

func deleteWordAndTranslationAfterLastSending(id string, word string, translation string) {
	_, _ = db.Exec(
		fmt.Sprintf(`
		DELETE FROM %s 
		WHERE word='%s' AND translation='%s' AND yet_need_to_send=FALSE
		`,
			"user_"+id, word, translation))
}

func getWordAndUpdateYetNeedToSend(userIdStr string, word string, translation string, wordToSend utils.WordAndTranslation, rawName string) utils.WordAndTranslation {
	var e error
	if rawName != "" {
		_, e = db.Exec(
			fmt.Sprintf(`UPDATE %s SET %s=FALSE WHERE word='%s' AND translation='%s'`,
				"user_"+userIdStr,
				rawName,
				word,
				translation))
	}
	wordToSend = utils.WordAndTranslation{Word: word, Translation: translation}
	if e != nil {
		log.Panic(e)
	}

	return wordToSend
}
