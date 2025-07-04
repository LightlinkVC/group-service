package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/lightlink/group-service/internal/group/domain/dto"
	"github.com/lightlink/group-service/internal/group/domain/entity"
	"github.com/lightlink/group-service/internal/group/usecase"
)

type GroupHandler struct {
	groupUC usecase.GroupUsecaseI
}

func NewGroupHandler(groupUsecase usecase.GroupUsecaseI) *GroupHandler {
	return &GroupHandler{
		groupUC: groupUsecase,
	}
}

func (h *GroupHandler) StartCall(w http.ResponseWriter, r *http.Request) {
	groupIDString := mux.Vars(r)["groupID"]
	userIDString := r.Header.Get("X-User-ID")

	err := h.groupUC.StartCall(userIDString, groupIDString)
	if err != nil {
		fmt.Println("ERR: Error starting call")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func generateUserToken(secret, userID, roomID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		/*TODO Fix time*/
		"exp": time.Now().Add(time.Hour * 10).Unix(),
		"channels": []string{
			entity.RoomChannel(roomID),
			entity.UserChannel(roomID, userID),
			entity.GroupChannel(roomID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userIDString := r.Header.Get("X-User-ID")
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req dto.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}

	if len(req.Members) == 0 {
		http.Error(w, "Group must have at least one member", http.StatusBadRequest)
		return
	}

	groupEntity := &entity.Group{
		Name:      req.Name,
		CreatorID: uint(userID),
		TypeName:  "group",
	}

	var groupMembers []entity.GroupMember

	groupMembers = append(groupMembers, entity.GroupMember{
		UserID: uint(userID),
		Role:   "admin",
	})

	for _, m := range req.Members {
		groupMembers = append(groupMembers, entity.GroupMember{
			UserID: m.UserID,
			Role:   m.Role,
		})
	}

	if err := h.groupUC.Create(groupEntity, groupMembers); err != nil {
		http.Error(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":    groupEntity.Name,
		"groupID": groupEntity.ID,
	})
}

func (h *GroupHandler) GetGroups(w http.ResponseWriter, r *http.Request) {
	userIDString := r.Header.Get("X-User-ID")
	userID64, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	userID := uint(userID64)

	groups, err := h.groupUC.GetGroupsByUserID(userID)
	if err != nil {
		/*Handle*/
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	groupDTOs := []dto.GetGroupResponse{}
	for _, group := range groups {
		groupDTO := dto.GetGroupResponse{
			GroupID:   group.ID,
			GroupName: group.Name,
		}
		groupDTOs = append(groupDTOs, groupDTO)
	}

	response, err := json.Marshal(groupDTOs)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		fmt.Println("Failed to write groups response")
	}
}

func (h *GroupHandler) InfoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling incoming info request")
	userIDString := r.Header.Get("X-User-ID")
	groupID := mux.Vars(r)["groupID"]

	token, err := generateUserToken(os.Getenv("TOKEN_KEY"), userIDString, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"channels": map[string]string{
			"room":           entity.RoomChannel(groupID),
			"group_messages": entity.GroupChannel(groupID),
			"user":           entity.UserChannel(groupID, userIDString),
		},
	})
}

func (h *GroupHandler) GetPersonalGroupID(w http.ResponseWriter, r *http.Request) {
	userIDString := r.Header.Get("X-User-ID")
	userID64, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	userID := uint(userID64)

	friendIDString := mux.Vars(r)["friendID"]
	friendID64, err := strconv.ParseUint(friendIDString, 10, 32)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	friendID := uint(friendID64)

	groupID, err := h.groupUC.GetPersonalGroupID(userID, friendID)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	groupIDResponse := dto.GetPersonalGroupIDResponse{
		GroupID: groupID,
	}

	response, err := json.Marshal(groupIDResponse)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		fmt.Println("Failed to write get personal group id response")
	}
}

// func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		/*Handle*/
// 		fmt.Println(err)
// 		return
// 	}
// 	defer func() {
// 		if err = r.Body.Close(); err != nil {
// 			/*Handle*/
// 			fmt.Println(err)
// 		}
// 	}()

// 	createGroupRequest := dto.CreateGroupRequest{}
// 	err = json.Unmarshal(body, &createGroupRequest)
// 	if err != nil {
// 		/*Handle*/
// 		fmt.Println(err)
// 		return
// 	}

// 	userIDString := r.Header.Get("X-User-ID")
// 	userID64, err := strconv.ParseUint(userIDString, 10, 32)
// 	if err != nil {
// 		/*Handle*/
// 		fmt.Println(err)
// 		return
// 	}

// 	userID := uint(userID64)

// 	createGroupRequest.UserID = userID

// 	err = h.groupUC.Create(&createGroupRequest)
// 	if err != nil {
// 		/*Handle*/
// 		fmt.Println(err)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// }
