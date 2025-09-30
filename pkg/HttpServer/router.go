package HttpServer

import (
	"simple_blockchain/pkg/handler"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	CoreRouter *chi.Mux
}

func newRouter(handler *handler.Handler) *Router {
	router := chi.NewRouter()

	router.Route("/api", func(r chi.Router) {
		r.Get("/chain", handler.GetBlockchain)
		r.Post("/mine", handler.MineBlock)

		r.Get("/balance", handler.GetBalance)
		r.Get("/blocks", handler.GetAllBlocks)

		r.Get("/txs", handler.GetTransactions)
		r.Post("/tx/send", handler.SendTransaction)
		r.Get("/tx/fee", handler.GetCurrentTxFee)

		r.Post("/keys", handler.GenerateKeys)

		r.Delete("/clear", handler.ClearDatabase)

		r.Post("/node/register", handler.RegisterNode)
	})

	return &Router{
		CoreRouter: router,
	}
}
