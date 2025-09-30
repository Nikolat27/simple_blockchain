package handler

import (
	"fmt"
	"net/http"
	"simple_blockchain/pkg/utils"
)

func (handler *Handler) RegisterNode(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TcpAddress string `json:"tcp_address"`
	}

	if err := utils.ParseJSON(r, 10_000, &input); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	fmt.Println(input)

}
