package handler

import (
	"net/http"

	"github.com/gofiber/adaptor/v2"
	"ml-topup-v2/appsetup"
)

var app = appsetup.Setup()

func Handler(w http.ResponseWriter, r *http.Request) {
	adaptor.FiberApp(app)(w, r)
}
