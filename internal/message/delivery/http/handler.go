package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/centrifugal/centrifuge"
	"github.com/gorilla/mux"
	"github.com/lightlink/group-service/internal/message/domain/dto"
	"github.com/lightlink/group-service/internal/message/usecase"
)

type MessageHandler struct {
	messageUC usecase.MessageUsecaseI
	node      *centrifuge.Node
}

func NewMessageHandler(messageUC usecase.MessageUsecaseI, node *centrifuge.Node) *MessageHandler {
	return &MessageHandler{
		messageUC: messageUC,
		node:      node,
	}
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		/*Handle*/
		fmt.Println("Failed to read request body")
		return
	}
	defer func() {
		if err = r.Body.Close(); err != nil {
			/*Handle*/
			fmt.Println("Failed to close request body")
		}
	}()

	createMessageRequest := &dto.CreateMessageRequest{}
	err = json.Unmarshal(body, createMessageRequest)
	if err != nil {
		/*Handle*/
		fmt.Println("failed unmarshal")
		return
	}

	userIDString := r.Header.Get("X-User-ID")
	userID64, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	userID := uint(userID64)
	createMessageRequest.UserID = userID

	createdMessage, err := h.messageUC.Create(createMessageRequest)
	if err != nil {
		/*Handle*/
		fmt.Println("failed create message")
		return
	}

	response, err := json.Marshal(createdMessage)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	_, err = h.node.Publish(
		fmt.Sprintf("group:%d", createMessageRequest.GroupID),
		response,
	)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MessageHandler) GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	groupIDString := mux.Vars(r)["groupID"]
	groupID64, err := strconv.ParseUint(groupIDString, 10, 32)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	groupID := uint(groupID64)

	messages, err := h.messageUC.GetByGroupID(groupID)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	response, err := json.Marshal(messages)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		fmt.Println("Failed to write get messages response")
	}
}
