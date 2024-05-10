package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yafireyhan01/e-wallet/model"
	"github.com/yafireyhan01/e-wallet/model/dto"
	"github.com/yafireyhan01/e-wallet/usecase"
	"github.com/yafireyhan01/e-wallet/utils/common"
)

type TopupController struct {
	ut usecase.TopupUseCase
	uc usecase.UserUseCase
	rg *gin.RouterGroup
}

// tulis handler code kalian disini
func (t *TopupController) CreateTopupHandler(c *gin.Context) {
	var payload model.TopupModel
	var ammount dto.TopupRequest
	c.ShouldBind(&ammount)
	payload.TransactionDetails.GrossAmt = int64(ammount.Ammount)
	claims, exists := c.Get("claims")
	if !exists {
		common.SendErrorResponse(c, http.StatusBadRequest, "Sepertinya login anda tidak valid")
		return
	}
	payload.User.Id = claims.(*common.JwtClaim).DataClaims.Id
	payload.User, _ = t.uc.FindById(payload.User.Id)
	res, err := t.ut.CreateTopup(payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.SendCreateResponse(c, "SUCCESS", res)
}

func (t *TopupController) ResponseTopupHandler(c *gin.Context) {
	var payload dto.ResponsePayment
	payload.OrderId = c.Query("order_id")
	payload.StatusCode, _ = strconv.Atoi(c.Query("status_code"))
	payload.TransactionStatus = c.Query("transaction_status")

	res, err := t.ut.PaymentUpdate(payload)
	if err != nil {
		common.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	common.SendSingleResponse(c, "SUCCESS", res)
}

func (t *TopupController) HistoryTopupHandler(c *gin.Context) {
	var id string
	var page int
	claims, exists := c.Get("claims")
	if !exists {
		common.SendErrorResponse(c, http.StatusBadRequest, "Sepertinya login anda tidak valid")
		return
	}
	id = claims.(*common.JwtClaim).DataClaims.Id
	page, _ = strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	datas, err := t.ut.FindAll(id, page)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.SendSingleResponse(c, "SUCCESS", datas)
}

func (t *TopupController) HistoryAdminTopupHandler(c *gin.Context) {
	var id string
	var page int
	id = c.Param("id")
	page, _ = strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}

	datas, err := t.ut.FindAll(id, page)
	if err != nil {
		common.SendErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.SendSingleResponse(c, "SUCCESS", datas)
}

func (t *TopupController) Route() {
	rg := t.rg.Group("/topup")
	{
		// tulis route disini
		rg.POST("/", common.JWTAuth("user"), t.CreateTopupHandler)
		rg.GET("/response", t.ResponseTopupHandler)
		rg.GET("/history", common.JWTAuth("user"), t.HistoryTopupHandler)
		rg.GET("/history/:id", common.JWTAuth("admin"), t.HistoryAdminTopupHandler)
	}
}

func NewTopupController(ut usecase.TopupUseCase, uc usecase.UserUseCase, rg *gin.RouterGroup) *TopupController {
	return &TopupController{ut: ut, uc: uc, rg: rg}
}
