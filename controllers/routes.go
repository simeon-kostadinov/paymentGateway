package controllers

import "github.com/xectich/paymentGateway/middlewares"

//initializeRoute: used when creating the server to init the routes
func (s *Server) initializeRoutes() {
	// Login Route
	s.Router.HandleFunc("/login", middlewares.SetMiddlewareJSON(s.Login)).Methods("POST")

	//Authorization routes
	s.Router.HandleFunc("/{mid}/authorize", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.RequestAuthorization))).Methods("PUT")
	s.Router.HandleFunc("/{mid}/capture", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.Capture))).Methods("POST")
	s.Router.HandleFunc("/{mid}/void", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.Void))).Methods("POST")
	s.Router.HandleFunc("/{mid}/refund", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.Refund))).Methods("POST")
}
