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

	// Waits for blockchain data from the node`s channel
	blocks := <-handler.Node.BlockchainRespCh

	valid, err := handler.Node.Blockchain.VerifyBlocks(blocks)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	if !valid {
		utils.WriteJSON(w, http.StatusBadRequest, "received blockchain is corrupted")
		return
	}

	handler.Node.AddNewPeer(input.TcpAddress)

	utils.WriteJSON(w, http.StatusAccepted, "Blockchain verified successfully!")
}

func (handler *Handler) GetPeers(w http.ResponseWriter, r *http.Request) {
	rows, err := handler.Node.Blockchain.Database.GetPeers()
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var peers []string
	for rows.Next() {
		var peer string
		if err := rows.Scan(&peer); err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, err)
			return
		}

		peers = append(peers, peer)
	}

	utils.WriteJSON(w, http.StatusOK, peers)
}
