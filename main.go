package main

import (
	chatbot "ai-chat-bot/internal/handler"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("hello world")

	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file:%v", err)
		return
	}

	r := gin.Default()
	r.POST("/dialogflow", chatbot.Handler)
	r.Run(":8090")
}
