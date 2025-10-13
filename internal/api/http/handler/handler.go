package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	StatusErr          = "error"
	StatusSuccess      = "success"
	StatusNotAvailable = "not available"
	StatusNotPermitted = "not permitted"
	StatusForbidden    = "forbidden"
)

// ResponseWithData
// @Description Общий ответ success/error, содержащий произвольные данные.
type ResponseWithData struct {
	Status string `json:"status"` // Результат запроса
	Data   any    `json:"data"`   // Объект полезной нагрузки
} // @Name _ResponseWithData

// ResponseWithMetaAndData
// @Description Общий ответ, который возвращает данные плюс разбивку на страницы или дополнительные метаданные.
type ResponseWithMetaAndData struct {
	Status   string `json:"status"`    // Результат запроса
	Data     any    `json:"data"`      // Объект полезной нагрузки
	Metadata any    `json:"_metadata"` // Метаданные
} // @Name _ResponseWithMetaAndData

// ResponseWithMessage
// @Description Общий простой ответ, который передает только понятное для человека сообщение.
type ResponseWithMessage struct {
	Status  string `json:"status"` // Результат запроса
	Message string `son:"message"` // Человеко-читаемое сообщение
} // @Name _ResponseWithMessage

// PaginationMetadata
// @Description Пагинация в стиле Offset/limit.
type PaginationMetadata struct {
	Page       int `example:"1"            json:"page"`       // Индекс текущей страницы (1 - базово)
	PageSize   int `example:"20"           json:"pageSize"`   // Элементов на каждой странице
	PageCount  int `example:"10"           json:"pageCount"`  // Общее количество страниц
	TotalCount int `example:"200"          json:"totalCount"` // Общее количество элементов
} // @Name _PaginationMetadata

func NoMethod(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, ResponseWithMessage{
		Status:  StatusNotAvailable,
		Message: "method not allowed on this endpoint",
	})
}

func NoRoute(c *gin.Context) {
	c.JSON(http.StatusNotFound, ResponseWithMessage{
		Status:  StatusNotAvailable,
		Message: "page not found",
	})
}
