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
		r.Post("/add/tx", handler.AddTransaction)
		r.Post("/mine", handler.MineBlock)
		r.Get("/balance", handler.GetBalance)

		r.Get("/transactions", handler.GetTransactions)

		r.Post("/keys", handler.GenerateKeys)

		r.Post("/node/register", handler.RegisterNode)
	})

	return &Router{
		CoreRouter: router,
	}
}
