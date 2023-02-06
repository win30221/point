package delivery

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/win30221/core/http/ctx"
	"github.com/win30221/core/http/response"
	"github.com/win30221/core/http/validate"
	"github.com/win30221/point/domain"
	"github.com/win30221/point/service/usecase"
)

type Delivery struct {
	useCase *usecase.UseCase
}

func NewDelivery(r *gin.RouterGroup, useCase *usecase.UseCase) {
	d := &Delivery{
		useCase: useCase,
	}

	r.POST("/wallets", d.CreateWallet)
	r.GET("/wallets/:userID", d.GetPointBalance)
	r.PATCH("/wallets/:userID", d.UpdatePoint)

	r.GET("/logs/:userID", d.GetUserPointLogs)
}

// CreateWallet godoc
// @Summary 新增積分錢包
// @Security Systoken
// @Description 使用模組: user
// @Param - formData domain.CreateWalletReq true " "
// @Success 200 {string} string
// @Router /point/wallets [post]
func (d *Delivery) CreateWallet(c *gin.Context) {
	ctx := ctx.New(c, c.Request.Context())
	req := domain.CreateWalletReq{}

	err := c.ShouldBind(&req)
	if err != nil {
		response.BindParameterError(ctx, err)
		return
	}

	err = validate.Struct(req)
	if err != nil {
		response.ValidParameterError(ctx, err)
		return
	}

	err = d.useCase.CreateWallet(ctx, req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err)
		return
	}

	response.OK(ctx, nil)
}

// GetPointBalance godoc
// @Summary 取得積分錢包餘額
// @Security Systoken
// @Description 使用模組: user-profile
// @Param UserID path string true " "
// @Success 200 {integer} float64 "用戶積分錢包餘額"
// @Router /point/wallets/{userID} [get]
func (d *Delivery) GetPointBalance(c *gin.Context) {
	ctx := ctx.New(c, c.Request.Context())

	userID := c.Param("userID")
	if len(userID) == 0 {
		err := errors.New("userID required")
		response.ValidParameterError(ctx, err)
		return
	}

	balance, err := d.useCase.GetBalance(ctx, userID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err)
		return
	}

	response.OK(ctx, balance)
}

// UpdatePoint godoc
// @Summary 變更積分
// @Security Systoken
// @Description 使用模組: fortune-recommend, mission, user, reward
// @Param - formData domain.UpdatePointReq true " "
// @Param UserID path string true " "
// @Success 200 {object} domain.UpdatePointRes
// @Router /point/wallets/{userID} [patch]
func (d *Delivery) UpdatePoint(c *gin.Context) {
	ctx := ctx.New(c, c.Request.Context())
	req := domain.UpdatePointReq{}

	err := c.ShouldBind(&req)
	if err != nil {
		response.BindParameterError(ctx, err)
		return
	}

	err = validate.Struct(req)
	if err != nil {
		response.ValidParameterError(ctx, err)
		return
	}

	userID := c.Param("userID")
	if userID == "" {
		err = errors.New("userID required")
		response.ValidParameterError(ctx, err)
		return
	}

	balance, err := d.useCase.UpdatePoint(ctx, userID, req.OpCode, req.Type, req.Difference)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err)
		return
	}

	res := domain.UpdatePointRes{
		Balance: balance,
	}

	response.OK(ctx, res)
}

// GetUserPointLogs godoc
// @Summary 取得用戶積分異動紀錄
// @Security Systoken
// @Description 使用模組: user-profile
// @Param UserID path string true " "
// @Success 200 {object} []domain.PointLog "用戶積分異動紀錄"
// @Router /point/logs/{userID} [get]
func (d *Delivery) GetUserPointLogs(c *gin.Context) {
	ctx := ctx.New(c, c.Request.Context())

	userID := c.Param("userID")
	if userID == "" {
		err := errors.New("userID required")
		response.ValidParameterError(ctx, err)
		return
	}

	balance, err := d.useCase.GetUserPointLogs(ctx, userID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err)
		return
	}

	response.OK(ctx, balance)
}
