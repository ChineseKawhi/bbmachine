package bbservice

import (
	"bbmachine/connection"
	"bbmachine/utils"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

//Db数据库连接池

type Service struct {
	config *viper.Viper
	sf     utils.Snowflake

	Cm connection.ConnectionManager
}

func NewService() Service {

	viper := viper.New()
	viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")

	// any approach to require this configuration into your program.
	absConfigPath, err := filepath.Abs("./config")
	if err != nil {
		return Service{}
	}
	viper.AddConfigPath(absConfigPath)
	viper.ReadInConfig()

	err = InitDb(viper)
	if err != nil {
		fmt.Printf("err in InitDb: %v", err)
		return Service{}
	}

	return Service{sf: utils.Snowflake{}, config: viper, Cm: connection.NewConnectionManager()}
}

func (s *Service) GetCreateUser(userName string) User {
	user := User{}
	err := user.Get(userName)
	if err != nil {
		fmt.Printf("err in GetUser: %v\n", err)
	}
	if user.UserId == 0 {
		user.UserId = s.sf.Next().Int64()
		user.UserName = userName
		err = user.Insert()
		if err != nil {
			fmt.Printf("err in InsertUser: %v\n", err)
		}
	}
	return user
}

func (s *Service) SendMessage(sender User, chat Chat, body string) error {

	chatters := s.GetChatters(chat.ChatId)

	msg := Message{sender.UserId, s.sf.Next().Int64(), chat.ChatId, body}
	err := msg.Insert()
	if err != nil {
		return err
	}

	// var inboxes InBoxes
	for _, chatter := range chatters {
		if chatter.UserId != sender.UserId {
			go func(receiverId int64) {
				is_sent := false
				conn := s.Cm.Get(receiverId)
				if conn != nil {
					for conn.WritingDb {
						fmt.Printf("wait for %v\n", receiverId)
						time.Sleep(1 * time.Second)
					}
					err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("In %v: %v said: %v", chat.ChatName, sender.UserName, msg.Body)))
					if err != nil {
						fmt.Printf("write in conn %v: %v", chatter.UserId, err)
					}
					is_sent = true
				}
				inbox := InBox{UserId: receiverId, SenderId: sender.UserId, ChatId: chat.ChatId, MessageId: msg.MessageId, Body: msg.Body, IsSent: is_sent}
				// inboxes = append(inboxes, InBox{UserId: receiverId, SenderId: senderId, ChatId: chat.ChatId, MessageId: msg.MessageId, Body: msg.Body, IsSent: is_sent})
				err := inbox.Insert()
				if err != nil {
					fmt.Printf("err in box insert is %v", err)
				}
			}(chatter.UserId)
		}
	}
	return nil
}

func (s *Service) ReceiveMessage(userId int64, conn *connection.Connection) error {
	var inboxes InBoxes
	err := inboxes.Select(userId)
	if err != nil {
		return err
	}
	for _, inbox := range inboxes {
		err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("In %v: %v said: %v", inbox.ChatName, inbox.SenderName, inbox.Body)))
		if err != nil {
			fmt.Printf("write in conn %v: %v", inbox.UserId, err)
		}
		inbox.IsSent = true
	}
	inboxes.Save()
	return nil
}

func (s *Service) NewChat(creatorId int64, chatterIds []int64) Chat {
	chat := Chat{ChatId: s.sf.Next().Int64(), CreatedBy: creatorId}
	err := chat.Insert()
	if err != nil {
		fmt.Printf("err in AddChat: %v\n", err)
	}
	chatters := make(Chatters, len(chatterIds))
	for i, chatter := range chatterIds {
		chatters[i] = Chatter{UserId: chatter, ChatId: chat.ChatId}
	}
	err = chatters.BulkInsert()
	if err != nil {
		fmt.Printf("err in AddChatters: %v\n", err)
	}
	return chat
}

func (s *Service) GetChat(chatName string) Chat {
	chat := Chat{}
	err := chat.Get(chatName)
	if err != nil {
		fmt.Printf("err in GetChat: %v\n", err)
	}
	return chat
}

func (s *Service) GetCreateChat(createUserId int64, chatName string) Chat {
	chat := Chat{}
	err := chat.Get(chatName)
	if err != nil {
		fmt.Printf("err in GetChat: %v\n", err)
	}
	if chat.ChatId == 0 {
		chat.ChatId = s.sf.Next().Int64()
		chat.ChatName = chatName
		chat.CreatedBy = createUserId
		err = chat.Insert()
		if err != nil {
			fmt.Printf("err in InsertChat: %v\n", err)
		}
	}
	return chat
}

func (s *Service) GetChatters(chatId int64) []Chatter {
	chatters := Chatters{}
	err := chatters.Select(chatId)
	if err != nil {
		fmt.Printf("err in GetChatters: %v\n", err)
	}
	return chatters
}

func (s *Service) GetCreateChatter(userId int64, chatId int64) Chatter {
	chatter := Chatter{}
	err := chatter.Get(userId, chatId)
	if err != nil {
		fmt.Printf("err in GetChatter: %v\n", err)
	}
	if chatter.UserId == 0 {
		chatter.UserId = userId
		chatter.ChatId = chatId
		err = chatter.Insert()
		if err != nil {
			fmt.Printf("err in InsertChatter: %v\n", err)
		}
	}
	return chatter
}
