package domain

import "time"

type PointLog struct {
	UserID     string    `json:"userID"`
	OpCode     string    `json:"opCode"`
	Type       string    `json:"type"`
	Before     float64   `json:"before"`
	Difference float64   `json:"difference"`
	CreatedAt  time.Time `json:"createdAt"`
}

// -------- 以下為 http request 及 response 結構

type CreateWalletReq struct {
	UserID []string `form:"userID" validate:"gt=0,dive,required"`
}

type UpdatePointReq struct {
	Difference float64 `form:"difference" validate:"required"`
	OpCode     string  `form:"opCode" validate:"required"`
	Type       string  `form:"type" validate:"required"`
}

type UpdatePointRes struct {
	Balance float64 `json:"balance"`
}
