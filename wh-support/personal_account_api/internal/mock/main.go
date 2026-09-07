//nolint:all
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type PhotoInfo struct {
	Photo []byte `json:"photo_base64"`
}

type EmployeePhotoResponse struct {
	Data PhotoInfo `json:"data"`
}

type EmployeeInfo struct {
	ID   int64  `json:"employee_id"`
	Name string `json:"employee_name"`

	OfficeID   *int64  `json:"office_id"`
	OfficeName *string `json:"office_name"`

	WhID   *int64  `json:"wh_id"`
	WhName *string `json:"wh_name"`

	IsDeleted bool `json:"isdeleted"`
}

type EmployeeInfoResp struct {
	Data EmployeeInfo `json:"data"`
}

func handle_photo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, EmployeePhotoResponse{Data: PhotoInfo{
		Photo: []byte("ijrkfjhergfj=="),
	}})
}

func handle_employee_info(ctx *gin.Context) {
	officeID := int64(123)
	whID := int64(123)
	officeName := "office"
	whName := "wh"
	ctx.JSON(http.StatusOK, EmployeeInfoResp{Data: EmployeeInfo{
		ID:         123,
		Name:       "Pupa Lupa Haha",
		OfficeID:   &officeID,
		OfficeName: &officeName,
		WhID:       &whID,
		WhName:     &whName,
		IsDeleted:  false,
	}})
}

func main() {
	e := gin.Default()
	e.Handle(http.MethodGet, "/api/employee/info", handle_employee_info)
	e.Handle(http.MethodGet, "/api/employee/photo", handle_photo)

	server := &http.Server{
		Addr:    fmt.Sprintf(":8084"),
		Handler: e,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
