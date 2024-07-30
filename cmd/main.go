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
		port := viper.GetString("port")
		host := viper.GetString("host")
		addr := host + ":" + port
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			logrus.Fatalf(err.Error())
		}
	}()

	wg.Wait()

}

func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
