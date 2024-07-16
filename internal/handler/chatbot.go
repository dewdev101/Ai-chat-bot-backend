package chatbot

import (
	"context"
	"fmt"
	"net/http"
	"os"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"cloud.google.com/go/dialogflow/apiv2/dialogflowpb"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
)

func Handler(c *gin.Context) {

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectID := os.Getenv("DIALOGFLOW_PROJECT_ID")
	sessionID := requestData.SessionID
	if sessionID == "" {
		sessionID = "default_session"
	}
	text := requestData.Text
	if text == "" {
		text = "Hello"
	}

	sessionPath := "projects/" + projectID + "/agent/sessions/" + sessionID
	ctx := c.Request.Context()

	sessionClient, err := dialogflow.NewSessionsClient(ctx, option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_CREDENTIALS_JSON"))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer sessionClient.Close()

	textInput := &dialogflowpb.TextInput{Text: text, LanguageCode: "en-US"}
	queryInput := &dialogflowpb.QueryInput{Input: &dialogflowpb.QueryInput_Text{Text: textInput}}
	request := &dialogflowpb.DetectIntentRequest{
		Session:    sessionPath,
		QueryInput: queryInput,
	}

	response, err := sessionClient.DetectIntent(ctx, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	queryResult := response.GetQueryResult()
	intentName := queryResult.GetIntent().GetDisplayName()

	if intentName == "Default Fallback Intent" || intentName == "" {
		client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-3.5-turbo",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "Please provide general information or engage in a casual conversation about: " + text,
					},
				},
				MaxTokens:   150,
				Temperature: 0.7,
			},
		)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"fulfillmentText": resp.Choices[0].Message.Content})
		return
	}

	c.JSON(http.StatusOK, gin.H{"fulfillmentText": queryResult.GetFulfillmentText()})
}
