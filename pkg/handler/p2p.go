package handler

import (
	"net/http"
	"simple_blockchain/pkg/p2p"
	"simple_blockchain/pkg/utils"
)

func (handler *Handler) NodeJoinNetwork(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TcpAddress string `json:"tcp_address"`
	}

	if err := utils.ParseJSON(r, 10_000, &input); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	if input.TcpAddress == handler.Node.GetCurrentAddress() {
		utils.WriteJSON(w, http.StatusBadRequest, "your TCP address is same as this node`s address")
		return
	}

	senderMsg := p2p.NewMessage(p2p.JoinNetworkMsg, input.TcpAddress, nil)

	// sends a 'join network' message
	if err := handler.Node.Write(handler.Node.GetCurrentAddress(), senderMsg.Marshal()); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusAccepted, "Blockchain verified successfully!")
}

func (handler *Handler) GetPeers(w http.ResponseWriter, r *http.Request) {
	allPeers := handler.Node.Peers

	utils.WriteJSON(w, http.StatusOK, allPeers)
}
