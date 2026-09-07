//nolint:all
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type EmployeeInfo struct {
	EmployeeID   int64  `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	Phone        string `json:"phone"`
	IsDeleted    bool   `json:"is_deleted"`
}

type EmployeeAuthorisationResponse struct {
	Data EmployeeInfo `json:"data"`
}

func handle_login(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, EmployeeAuthorisationResponse{Data: EmployeeInfo{
		EmployeeID:   1305064,
		EmployeeName: "Melnikov Dmitriy",
		Phone:        "+79005553535",
		IsDeleted:    false,
	}})
}

func handle_send_code(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNoContent)
}

func handle_get_all_spruts(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, []string{"SPRUT-507", "SPRUT-686", "SPRUT-117501", "SPRUT-117986", "SPRUT-120762"})
}

type EmployeeInfoName struct {
	Data struct {
		Name string `json:"employee_name"`
	} `json:"data"`
}

func handleEmpInfo(ctx *gin.Context) {
	reqId := ctx.Request.URL.Query()

	id, _ := strconv.ParseInt(reqId.Get("id"), 10, 64)

	switch id {
	case int64(1156237):
		var res EmployeeInfoName
		res.Data.Name = "Маша"
		ctx.JSON(http.StatusOK, res)
	case 1:
		var res EmployeeInfoName
		res.Data.Name = "иван иван иван"
		ctx.JSON(http.StatusOK, res)
	case 1305064:
		var res EmployeeInfoName
		res.Data.Name = "Мельников Дмитрий"
		ctx.JSON(http.StatusOK, res)
	default:
		var res EmployeeInfoName
		res.Data.Name = "ияяяяан"
		ctx.JSON(http.StatusOK, res)
	}
}

func main() {
	e := gin.Default()
	e.Handle(http.MethodPost, "/api/send_code", handle_send_code)
	e.Handle(http.MethodPost, "/api/login", handle_login)
	e.Handle(http.MethodGet, "/api/get_all_spruts", handle_get_all_spruts)
	e.Handle(http.MethodGet, "/api/employee/info", handleEmpInfo)

	server := &http.Server{
		Addr:    fmt.Sprintf(":8084"),
		Handler: e,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
