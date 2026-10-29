package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/service"
	httpin "github.com/SantiagoSu15/Task-it-Chat-MS/internal/infra/inbound/http"
	wsin "github.com/SantiagoSu15/Task-it-Chat-MS/internal/infra/inbound/websocket"
	redisout "github.com/SantiagoSu15/Task-it-Chat-MS/internal/infra/outbound/redis"
	wsout "github.com/SantiagoSu15/Task-it-Chat-MS/internal/infra/outbound/websocket"
	redisinfra "github.com/SantiagoSu15/Task-it-Chat-MS/internal/infrastructure/redis"
)

type App struct {
	server      *http.Server
	hub         *wsout.Hub
	redisClient *goredis.Client
}

func NewApp() *App {
	hub := wsout.NewHub()
	redisClient := redisinfra.NewClientFromEnv()
	chatRepository := redisout.NewChatRepository(redisClient)
	messageRepository := redisout.NewMessageRepository(redisClient)
	sendMessage := service.NewSendMessageService(chatRepository, messageRepository, wsout.NewChatWebSocket(hub))
	getHistory := service.NewGetMessageHistoryService(chatRepository, messageRepository)
	chatHTTP := httpin.NewChatController(getHistory)
	chatSocket := wsin.NewChatSocketController(sendMessage, getHistory, hub)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	chatHTTP.Routes(router)
	router.GET("/ws/chat", gin.WrapH(chatSocket))

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	return &App{
		hub:         hub,
		redisClient: redisClient,
		server: &http.Server{
			Addr:              ":" + port,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (a *App) Run() error {
	if a.server == nil {
		return errors.New("HTTP server is not configured")
	}
	defer a.redisClient.Close()
	if err := a.redisClient.Ping(context.Background()).Err(); err != nil {
		return err
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-stop:
		a.hub.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return a.server.Shutdown(ctx)
	}
}
