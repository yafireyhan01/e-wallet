package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yafireyhan01/e-wallet/model"
	"github.com/yafireyhan01/e-wallet/model/dto"
	"github.com/yafireyhan01/e-wallet/usecase"
	"github.com/yafireyhan01/e-wallet/utils/common"
)

type UserController struct {
	uc usecase.UserUseCase
	rg *gin.RouterGroup
}

func (e *UserController) getHandler(c *gin.Context) {
	id := c.Param("id")

	response, err := e.uc.GetBalanceCase(id)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, "Error "+err.Error())
		return
	}

	common.SendSingleResponse(c, "OK", response)
}

func (e *UserController) createHandler(c *gin.Context) {
	var payload dto.UserRequestDto
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	payloadResponse, err := e.uc.CreateUser(payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.SendCreateResponse(c, "SUCCESS", payloadResponse)
}

func (u *UserController) loginHandler(c *gin.Context) {
	var payload dto.LoginRequestDto
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	loginData, err := u.uc.LoginUser(payload)
	if err != nil {
		if err.Error() == "1" {
			common.SendErrorResponse(c, http.StatusForbidden, "Password salah")
			return
		}
		common.SendErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}
	common.SendSingleResponse(c, "success", loginData)
}

func (u *UserController) CheckBalance(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		common.SendErrorResponse(c, http.StatusInternalServerError, "Claims jwt tidak ada!")
		return
	}
	id := claims.(*common.JwtClaim).DataClaims.Id
	response, err := u.uc.GetBalanceCase(id)
	if err != nil {
		if err.Error() == "1" {
			common.SendErrorResponse(c, http.StatusBadRequest, "Verifikasi akun anda terlebih dahulu untuk akses cek saldo")
			return
		}
		common.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	common.SendSingleResponse(c, "SUCCESS", response)
}

func (s *UserController) UpdateHandler(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		common.SendErrorResponse(c, http.StatusBadRequest, "sepertinya login anda tidak valid")
		return
	}
	id := claims.(*common.JwtClaim).DataClaims.Id
	var payload dto.UserRequestDto
	if err := c.BindJSON(&payload); err != nil {
		common.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updatedUser, err := s.uc.UpdateUser(id, payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	common.SendSingleResponse(c, "UPDATE SUCCESS", updatedUser)
}

func (u *UserController) VerifyHandler(c *gin.Context) {
	var payload dto.VerifyUser
	claims, exists := c.Get("claims")
	if !exists {
		common.SendErrorResponse(c, http.StatusInternalServerError, "Claims jwt tidak ada!")
		fmt.Println(2)
		return
	}
	payload, err := common.FileVerifyHandler(c)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, "failed upload photo"+err.Error())
		return
	}
	payload.UserId = claims.(*common.JwtClaim).DataClaims.Id

	response, err := u.uc.VerifyUser(payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	common.SendCreateResponse(c, "success", response)
}

func (p *UserController) UpdatePinHandler(c *gin.Context) {
	var payload dto.UpdatePinRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		common.SendErrorResponse(c, http.StatusInternalServerError, "Claims jwt tidak ada!")
		return
	}

	payload.UserId = claims.(*common.JwtClaim).DataClaims.Id
	payloadResponse, _ := p.uc.GetBalanceCase(payload.UserId)

	if payload.OldPin != payloadResponse.Pin {
		common.SendErrorResponse(c, http.StatusBadRequest, "Old pin not match")
		return
	}

	payload.OldPin = payloadResponse.Pin

	response, err := p.uc.UpdatePinUser(payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.SendSingleResponse(c, "success", response)
}

func (p *UserController) CreateRekeningHandler(c *gin.Context) {
	var payload model.Rekening
	err := c.ShouldBind(&payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		common.SendErrorResponse(c, http.StatusUnauthorized, "sepertinya login anda tidak valid")
		return
	}
	payload.UserId = claims.(*common.JwtClaim).DataClaims.Id
	res, err := p.uc.CreateRekening(payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.SendCreateResponse(c, "SUCCESS", res)
}

func (p *UserController) GetRekeningHandler(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		common.SendErrorResponse(c, http.StatusUnauthorized, "sepertinya login anda tidak valid")
		return
	}
	id := claims.(*common.JwtClaim).DataClaims.Id
	res, err := p.uc.FindRekening(id)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.SendSingleResponse(c, "SUCCESS", res)
}

func (p *UserController) Route() {
	p.rg.POST("/users/login", p.loginHandler)
	p.rg.POST("/users", p.createHandler)
	p.rg.GET("/users/:id", common.JWTAuth("admin"), p.getHandler)
	p.rg.GET("/users/saldo", common.JWTAuth("user"), p.CheckBalance)
	p.rg.PUT("/users", common.JWTAuth("user"), p.UpdateHandler)
	p.rg.POST("/users/verify", common.JWTAuth("user"), p.VerifyHandler)
	p.rg.PUT("/users/pin", common.JWTAuth("user"), p.UpdatePinHandler)
	p.rg.POST("/users/rekening", common.JWTAuth("user"), p.CreateRekeningHandler)
	p.rg.GET("/users/rekening", common.JWTAuth("user"), p.GetRekeningHandler)

}

func NewUserController(uc usecase.UserUseCase, rg *gin.RouterGroup) *UserController {
	return &UserController{
		uc: uc,
		rg: rg,
	}
}
