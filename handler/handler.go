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
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	Env     string
	Path    string
	MaxSize uint64 `mapstructure:"max_size"`
	Cron    string
	Level   string
}

type Handler struct {
	Config  *viper.Viper
	service *bbservice.Service
	logger  *zap.Logger
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

	var cfg LogConfig
	if err := viper.UnmarshalKey("logging", &cfg); err != nil {
		panic(err)
	}
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	absPath, err := filepath.Abs(cfg.Path)
	fmt.Printf("%v", absPath)
	if err != nil {
		panic(err)
	}
	rotateWriter := &lumberjack.Logger{
		Filename: absPath,
		MaxSize:  int(cfg.MaxSize), // MB
	}
	level, err := strToLevel(cfg.Level, zapcore.DebugLevel)
	if err != nil {
		panic(err)
	}
	core := zapcore.NewCore(encoder, zapcore.AddSync(rotateWriter), zap.NewAtomicLevelAt(level))
	logger := zap.New(core)
	if err != nil {
		panic(err)
	}
	// logger := zap.NewExample()

	return Handler{service: bbservice.NewService(), Config: viper, logger: logger}
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
	h.logger.Info("login", zap.String("user", req.UserName))

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Upgrade", zap.Error(err))
		return err
	}

	conn := connection.Connection{Conn: c, WritingDb: true}
	user := h.service.GetCreateUser(req.UserName)

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
				stream := strings.Split(string(message), "|")
				chat := h.service.GetCreateChat(user.UserId, stream[0])
				body := strings.Join(stream[1:], "")
				_ = h.service.GetCreateChatter(user.UserId, chat.ChatId)
				err := h.service.SendMessage(user, chat, body)
				if err != nil {
					fmt.Printf("err in send is: %v", err)
				}
				log.Printf("recv: %s", message)
			}()
		}
	}()
	return nil
}

func strToLevel(level string, defaultLevel zapcore.Level) (zapcore.Level, error) {
	if len(level) != 0 {
		if err := defaultLevel.UnmarshalText([]byte(level)); err != nil {
			return zapcore.FatalLevel, fmt.Errorf("unsupported logging level %s", level)
		}
	}
	return defaultLevel, nil
}
