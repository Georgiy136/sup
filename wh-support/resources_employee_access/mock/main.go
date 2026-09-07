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
	Data struct {
		Name string `json:"employee_name"`
	} `json:"data"`
}

func handleEmpInfo(ctx *gin.Context) {
	reqId := ctx.Request.URL.Query()

	id, _ := strconv.ParseInt(reqId.Get("id"), 10, 64)

	switch id {
	case int64(1305064):
		var res EmployeeInfo
		res.Data.Name = "123 1231 123"
		ctx.JSON(http.StatusOK, res)
	case 1:
		var res EmployeeInfo
		res.Data.Name = "иван иван иван"
		ctx.JSON(http.StatusOK, res)
	default:
		var res EmployeeInfo
		res.Data.Name = "ияяяяан"
		ctx.JSON(http.StatusOK, res)
	}
}

func main() {
	e := gin.Default()
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
