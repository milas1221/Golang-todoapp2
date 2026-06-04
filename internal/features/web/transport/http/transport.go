// Package web_transport_http содержит HTTP-обработчики для раздачи веб-страниц.
package web_transport_http

import (
	"net/http"

	"github.com/milas1221/golang-todoapp/internal/core/domain"
	core_http_server "github.com/milas1221/golang-todoapp/internal/core/transport/http/server"
)

// WebHTTPHandler — HTTP-обработчик для статических страниц веб-интерфейса.
type WebHTTPHandler struct {
	webService WebService
}

// WebService — интерфейс сервиса для получения веб-страниц.
type WebService interface {
	GetMainPage() (domain.File, error)
}

// NewWebHTTPHandler создаёт обработчик веб-страниц с внедрённым сервисом.
func NewWebHTTPHandler(
	webService WebService,
) *WebHTTPHandler {
	return &WebHTTPHandler{
		webService: webService,
	}
}

// Routes возвращает маршрут главной страницы приложения.
// Регистрируется без API-префикса через httpServer.RegisterRoutes.
func (h *WebHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/",
			Handler: h.GetMainPage,
		},
	}
}
