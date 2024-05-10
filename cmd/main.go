package main

import (
	"GameService/connectteam_service/requests"
	"GameService/game"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"sync"
)

func main() {
	//if err := initConfig(); err != nil {
	//	logrus.Fatalf("error")
	//}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error")
	}

	httpService := requests.NewHTTPService(os.Getenv("HTTP_SERVICE_API_KEY"))

	wsServer := game.NewWebsocketServer(httpService)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		wsServer.Run()
	}()

	go func() {
		defer wg.Done()
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			game.ServeWs(wsServer, w, r)
		})
		http.ListenAndServe(":8080", nil)
	}()

	wg.Wait()

}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
