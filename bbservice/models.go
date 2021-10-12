package bbservice

import (
	"bbmachine/utils"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var DB *sqlx.DB

func InitDb(config *viper.Viper) error {
	var dbConfig utils.DBConfig
	if err := config.UnmarshalKey("database", &dbConfig); err != nil {
		return err
	}
	db, err := sqlx.Connect(dbConfig.Driver, dbConfig.Source)
	if err != nil {
		fmt.Printf("err in Connect: %v", err)
	}
	DB = db
	return nil
}

type User struct {
	UserId   int64  `db:"user_id"`
	UserName string `db:"user_name"`
}

func (user *User) Insert() error {
	now := time.Now().Unix()
	sql := fmt.Sprintf(`INSERT INTO user 
	(updated_at, created_at, user_id, user_name) values
	(?, ?, ?, ?)`)
	_, err := DB.Exec(sql, now, now, user.UserId, user.UserName)
	if err != nil {
		return err
	}
	return nil
}

func (user *User) Get(userName string) error {
	err := DB.Get(user, "SELECT user_id, user_name FROM user WHERE user_name=?", userName)
	if err != nil {
		return err
	}
	return nil
}

type Message struct {
	SenderId  int64  `db:"sender_id"`
	MessageId int64  `db:"message_id"`
	ChatId    int64  `db:"chat_id"`
	Body      string `db:"body"`
}

func (msg *Message) Insert() error {
	now := time.Now().Unix()
	sql := fmt.Sprintf(`INSERT INTO message 
	(updated_at, created_at, message_id, sender_id, chat_id, body) values
	(?, ?, ?, ?, ?, ?)`)
	_, err := DB.Exec(sql, now, now, msg.MessageId, msg.SenderId, msg.ChatId, msg.Body)
	if err != nil {
		return err
	}
	return nil
}

type Chat struct {
	UpdatedAt int64  `db:"updated_at"`
	ChatName  string `db:"chat_name"`
	CreatedAt int64  `db:"created_at"`
	ChatId    int64  `db:"chat_id"`
	CreatedBy int64  `db:"created_by"`
}

func (chat *Chat) Insert() error {
	now := time.Now().Unix()
	sql := fmt.Sprintf(`INSERT INTO chat 
		(updated_at, created_at, chat_name, chat_id, created_by) values
		(?, ?, ?, ?, ?)`)
	_, err := DB.Exec(sql, now, now, chat.ChatName, chat.ChatId, chat.CreatedBy)
	if err != nil {
		return err
	}
	return nil
}

func (chat *Chat) Get(chatName string) error {
	err := DB.Get(chat, "SELECT updated_at, created_at, chat_name, chat_id, created_by FROM chat WHERE chat_name=?", chatName)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", chat)
	return nil
}

type Chatter struct {
	UserId int64 `db:"user_id"`
	ChatId int64 `db:"chat_id"`
}
type Chatters []Chatter

func (chatter *Chatter) Insert() error {
	now := time.Now().Unix()
	sql := fmt.Sprintf(`INSERT INTO chatter 
	(updated_at, created_at, user_id, chat_id) values
	(?, ?, ?, ?)`)
	_, err := DB.Exec(sql, now, now, chatter.UserId, chatter.ChatId)
	if err != nil {
		return err
	}
	return nil
}

func (chatter *Chatter) Get(userId int64, chatId int64) error {
	err := DB.Get(chatter, "SELECT user_id, chat_id FROM chatter WHERE user_id=? and chat_id=?", userId, chatId)
	if err != nil {
		return err
	}
	return nil
}

func (chatters *Chatters) BulkInsert() error {
	now := time.Now().Unix()
	sql := fmt.Sprintf(`INSERT INTO chatter
		(updated_at, created_at, user_id, chat_id) values
		(?, ?, ?, ?)`)
	for _, chatter := range *chatters {
		_, err := DB.Exec(sql, now, now, chatter.UserId, chatter.ChatId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (chatters *Chatters) Select(chatId int64) error {

	err := DB.Select(chatters, "SELECT user_id, chat_id FROM chatter WHERE chat_id=?", chatId)
	if err != nil {
		fmt.Printf("err in GetChatters: %v\n", err)
		return err
	}
	return nil
}

type InBox struct {
	Id         int64  `db:"id"`
	UpdatedAt  int64  `db:"updated_at"`
	CreatedAt  int64  `db:"created_at"`
	UserId     int64  `db:"user_id"`
	SenderId   int64  `db:"sender_id"`
	SenderName string `db:"sender_name"`
	ChatId     int64  `db:"chat_id"`
	ChatName   string `db:"chat_name"`
	MessageId  int64  `db:"message_id"`
	Body       string `db:"body"`
	IsSent     bool   `db:"is_sent"`
}
type InBoxes []InBox

func (inBoxes *InBoxes) BulkInsert() error {
	now := time.Now().Unix()
	sql := fmt.Sprintf(`INSERT INTO inbox
		(updated_at, created_at, user_id, sender_id, chat_id, message_id, body, is_sent) values
		(?, ?, ?, ?, ?, ?, ?, ?)`)
	for _, inBox := range *inBoxes {
		_, err := DB.Exec(sql, now, now, inBox.UserId, inBox.SenderId, inBox.ChatId, inBox.MessageId, inBox.Body, inBox.IsSent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inBox *InBox) Insert() error {
	now := time.Now().Unix()
	sql := fmt.Sprintf(`INSERT INTO inbox
		(updated_at, created_at, user_id, sender_id, chat_id, message_id, body, is_sent) values
		(?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err := DB.Exec(sql, now, now, inBox.UserId, inBox.SenderId, inBox.ChatId, inBox.MessageId, inBox.Body, inBox.IsSent)
	if err != nil {
		return err
	}
	return nil
}

func (inBoxes *InBoxes) Select(userId int64) error {
	err := DB.Select(inBoxes, `SELECT id, a.updated_at, a.created_at, a.user_id, a.sender_id, a.chat_id, a.message_id, a.body, a.is_sent,
	b.user_name as sender_name, c.chat_name 
	FROM inbox a 
	left join user b on a.sender_id=b.user_id 
	left join chat c on a.chat_id=c.chat_id WHERE a.user_id=? and a.is_sent=0`, userId)
	if err != nil {
		fmt.Printf("err in GetInboxes: %v\n", err)
		return err
	}
	return nil
}

func (inBoxes *InBoxes) Save() error {
	sql := fmt.Sprintf(`Update inbox set is_sent=1 where id=?`)
	for _, inBox := range *inBoxes {
		_, err := DB.Exec(sql, inBox.Id)
		if err != nil {
			return err
		}
	}
	return nil
}
