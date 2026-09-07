//nolint:all
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
)

type EmployeeInfo struct {
	Data []Info `json:"data"`
}

type Info struct {
	ID   int64  `json:"employee_id"`
	Name string `json:"employee_name"`
}

func handleEmpInfo(ctx *gin.Context) {
	var req struct {
		IDs []int64 `json:"employee_ids"`
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to read body"})
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var res EmployeeInfo

	for _, id := range req.IDs {
		switch id {
		case 1305064:
			res.Data = append(res.Data, Info{ID: 1305064, Name: "123123123"})
		case 1:
			res.Data = append(res.Data, Info{ID: 1, Name: "ЯЯЯЯЯ"})
		case 546950:
			res.Data = append(res.Data, Info{ID: 546950, Name: "Уволен Увольвич"})
		}
	}

	if len(res.Data) == 0 {
		ctx.JSON(http.StatusNoContent, nil)
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func main() {
	e := gin.Default()
	e.Handle(http.MethodPost, "/api/employees_full_name", handleEmpInfo)

	server := &http.Server{
		Addr:    fmt.Sprintf(":8084"),
		Handler: e,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
