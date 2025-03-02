package controllers

import "github.com/go-chi/chi/v5"

type Controller interface {
	RegisterRoutes(router chi.Router)
}
