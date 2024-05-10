package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yafireyhan01/e-wallet/model"
	"github.com/yafireyhan01/e-wallet/model/dto"
	"github.com/yafireyhan01/e-wallet/usecase"
	"github.com/yafireyhan01/e-wallet/utils/common"
	"github.com/yafireyhan01/e-wallet/utils/encryption"
)

type AdminController struct {
	ua usecase.AdminUseCase
	uc usecase.UserUseCase
	rg *gin.RouterGroup
}

func (a *AdminController) RegisterHandler(c *gin.Context) {
	payload := model.Admin{}
	c.ShouldBind(&payload)
	payload.Password, _ = encryption.HashPassword(payload.Password)
	res, err := a.ua.RegisterAdmin(payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	common.SendSingleResponse(c, "SUCCESS", res)
}

func (a *AdminController) LoginHandler(c *gin.Context) {
	payload := dto.LoginRequestDto{}
	c.ShouldBind(&payload)

	response, err := a.ua.LoginAdmin(payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	common.SendSingleResponse(c, "SUCCESS", response)

}

func (a *AdminController) GetUserInfo(c *gin.Context) {
	userID := c.Param("id")

	user, err := a.ua.GetUserInfo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (a *AdminController) Route() {
	rg := a.rg.Group("/admin")
	{
		rg.POST("/", a.RegisterHandler)
		rg.POST("/login", a.LoginHandler)
		rg.GET("/user/:id", common.JWTAuth("admin"), a.GetUserInfo)
	}
}

func NewAdminController(ua usecase.AdminUseCase, uc usecase.UserUseCase, rg *gin.RouterGroup) *AdminController {
	return &AdminController{ua: ua, uc: uc, rg: rg}
}
