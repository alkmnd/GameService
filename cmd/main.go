package main

import (
	"GameService/game"
	"GameService/repository/requests"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
)

func main() {

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error")
	}

	httpService := requests.NewHTTPService(os.Getenv("HTTP_SERVICE_API_KEY"))

	zoomSDKKey := os.Getenv("ZOOM_SDK_KEY")
	zoomSDKSecret := os.Getenv("ZOOM_SDK_SECRET")
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
		http.ListenAndServe(":8080", nil)
		logrus.Println("ListenAndServe 8080")
	}()

	wg.Wait()

}
