package handler

import (
	"net/http"

	"github.com/gofiber/adaptor/v2"
	"ml-topup-v2/appsetup"
)

var app = appsetup.Setup()

func Handler(w http.ResponseWriter, r *http.Request) {
	// Debug
	println("[VERCEL_HANDLER] URL:", r.URL.String(), "RawQuery:", r.URL.RawQuery, "RequestURI:", r.RequestURI)
	
	// Fiber's adaptor parses r.RequestURI. In Vercel, r.RequestURI might not include RawQuery.
	if r.URL.RawQuery != "" {
		r.RequestURI = r.URL.Path + "?" + r.URL.RawQuery
	}
	
	adaptor.FiberApp(app)(w, r)
}
