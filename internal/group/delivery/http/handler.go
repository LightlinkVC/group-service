package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/lightlink/group-service/internal/group/domain/dto"
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

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}
	defer func() {
		if err = r.Body.Close(); err != nil {
			/*Handle*/
			fmt.Println(err)
		}
	}()

	createGroupRequest := dto.CreateGroupRequest{}
	err = json.Unmarshal(body, &createGroupRequest)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
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

	createGroupRequest.UserID = userID

	err = h.groupUC.Create(&createGroupRequest)
	if err != nil {
		/*Handle*/
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
