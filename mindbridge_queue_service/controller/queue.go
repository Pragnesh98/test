package controller

import (
	"fmt"
	"mindbridge_queue_service/interfaces"
	"net/http"

	"github.com/labstack/echo"
)

type QueueController struct {
	ucases interfaces.QueueUcaseInterface
}

func (r *QueueController) HandleQueue(c echo.Context) error {
	var t map[string]interface{}
	c.Bind(&t)
	ctx := c.Request().Context()
	fmt.Println("In exit queue ", c.Param("conference_uuid"), c.Param("conference_name"))
	authresp, err := r.ucases.HandleQueue(ctx, c.Param("conference_uuid"), c.Param("conference_name"))

	if err != nil {
		return c.JSON(http.StatusBadRequest, authresp)
	}
	return c.JSON(http.StatusOK, authresp)
}

func (r *QueueController) HandleEntryQueue(c echo.Context) error {
	var t map[string]interface{}
	c.Bind(&t)
	ctx := c.Request().Context()
	fmt.Println("In entry queue ", c.Param("conference_uuid"), c.Param("conference_name"))
	authresp, err := r.ucases.HandleEntryQueue(ctx, c.Param("conference_uuid"), c.Param("conference_name"))

	if err != nil {
		return c.JSON(http.StatusBadRequest, authresp)
	}
	return c.JSON(http.StatusOK, authresp)
}

func (r *QueueController) DeleteQueue(c echo.Context) error {
	var t map[string]interface{}
	c.Bind(&t)
	ctx := c.Request().Context()

	authresp, err := r.ucases.DeleteQueue(ctx)

	if err != nil {
		return c.JSON(http.StatusBadRequest, authresp)
	}
	return c.JSON(http.StatusOK, authresp)
}

func NewQueueController(e *echo.Echo, eusecase interfaces.QueueUcaseInterface) {
	hand := &QueueController{
		ucases: eusecase,
	}
	e.GET("/queue/:conference_uuid/:conference_name", hand.HandleQueue)
	e.GET("/queue/entry/:conference_uuid/:conference_name", hand.HandleEntryQueue)
	e.GET("/delete-queue", hand.DeleteQueue)
}
