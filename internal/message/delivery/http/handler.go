package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lightlink/group-service/internal/message/domain/dto"
	"github.com/lightlink/group-service/internal/message/usecase"
)

type MessageHandler struct {
	messageUC usecase.MessageUsecaseI
}

func NewMessageHandler(messageUC usecase.MessageUsecaseI) *MessageHandler {
	return &MessageHandler{
		messageUC: messageUC,
	}
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)

	groupIDStr := r.FormValue("group_id")
	groupID64, _ := strconv.ParseUint(groupIDStr, 10, 32)
	groupID := uint(groupID64)

	content := r.FormValue("content")

	files := r.MultipartForm.File["files"]

	createMessageRequest := dto.CreateMessageRequest{
		UserID:  userID,
		GroupID: groupID,
		Content: content,
		Files:   files,
	}

	message, err := h.messageUC.Create(&createMessageRequest)
	if err != nil {
		/*Handle*/
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("failed create message")
		return
	}

	response, err := json.Marshal(message)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		fmt.Println("Failed to write create message response")
	}
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
		w.WriteHeader(http.StatusBadRequest)
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
