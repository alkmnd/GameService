package main

import (
	"GameService/game"
	"GameService/repository/requests"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"sync"
)

func main() {

	if err := initConfig(); err != nil {
		logrus.Fatalf("error")
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error")
	}

	logrus.Println("Starting...")
	zoomSDKKey := os.Getenv("ZOOM_SDK_KEY")
	zoomSDKSecret := os.Getenv("ZOOM_SDK_SECRET")
	httpService := requests.NewHTTPService(os.Getenv("HTTP_SERVICE_API_KEY"),
		os.Getenv("ZOOM_API_ACCESS_TOKEN"),
		os.Getenv("ZOOM_API_REFRESH_TOKEN"), zoomSDKKey, zoomSDKSecret)
	generator := game.NewJWTGenerator(zoomSDKKey, zoomSDKSecret)

	wsServer := game.NewWebsocketServer(httpService, generator)
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
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			logrus.Fatalf(err.Error())
		}
		logrus.Println("ListenAndServe 8080")
	}()

	wg.Wait()

}

func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
