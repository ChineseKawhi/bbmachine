package handler

import (
	"bbmachine/bbservice"
	"bbmachine/connection"
	"bbmachine/proto"
	"bbmachine/utils"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type Handler struct {
	Config  *viper.Viper
	service bbservice.Service
}

func NewHandler() Handler {
	viper := viper.New()
	viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")

	// any approach to require this configuration into your program.
	absConfigPath, err := filepath.Abs("./config")
	if err != nil {
		return Handler{}
	}
	viper.AddConfigPath(absConfigPath)
	viper.ReadInConfig()

	return Handler{service: bbservice.NewService(), Config: viper}
}

func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) error {
	req := proto.CreateChatReq{}
	if err := utils.ReadHTTPReq(r, &req); err != nil {
		fmt.Print("parse fail")
		return err
	}

	chat := h.service.NewChat(req.Me, req.ChatterIds)
	fmt.Printf("chat is %+v", chat)

	res := &proto.CreateChatRes{ChatId: chat.ChatId, CreaterUserId: req.Me, ChatterIds: req.ChatterIds}
	return utils.WriteHTTPJSONRes(w, res)
}

// func JoinChat(w http.ResponseWriter, req *http.Request) {

// }

func (h *Handler) CatchErr(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	req := proto.InfoReq{}
	if err := utils.ReadHTTPReq(r, &req); err != nil {
		fmt.Print("parse InfoReq fail\n")
		return err
	}
	fmt.Printf("User login: %+v\n", req)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return err
	}

	conn := connection.Connection{Conn: c, WritingDb: true}
	user := h.service.GetCreateUser(req.UserName)
	_ = h.service.GetCreateChatter(user.UserId, 5577006791947779410)
	h.service.Cm.Set(user.UserId, &conn)
	h.service.ReceiveMessage(user.UserId, &conn)
	conn.WritingDb = false
	go func() {
		defer func() {
			c.Close()
			h.service.Cm.Del(user.UserId)
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			go func() {
				chat := h.service.GetChat(5577006791947779410)
				err := h.service.SendMessage(user, chat, string(message))
				if err != nil {
					fmt.Printf("err in send is: %v", err)
				}
				log.Printf("recv: %s", message)
			}()

			// err = c.WriteMessage(mt, message)
			// if err != nil {
			// 	log.Println("write:", err)
			// 	break
			// }
		}
	}()
	return nil
}
