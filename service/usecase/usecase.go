package usecase

import (
	"context"
	"fmt"

	"github.com/win30221/code/errno"
	"github.com/win30221/core/http/catch"
	"github.com/win30221/core/http/ctx"
	"github.com/win30221/point/domain"
	"github.com/win30221/point/service/repository"
	"go.uber.org/zap"
)

type UseCase struct {
	mysqlRepo *repository.MysqlRepo
	redisRepo *repository.RedisRepo
}

func NewUseCase(
	mysqlRepo *repository.MysqlRepo,
	redisRepo *repository.RedisRepo,
) *UseCase {
	return &UseCase{
		mysqlRepo: mysqlRepo,
		redisRepo: redisRepo,
	}
}

func (u *UseCase) CreateWallet(ctx ctx.Context, req domain.CreateWalletReq) (err error) {
	err = u.mysqlRepo.CreateWallet(ctx.Context, req)
	return
}

func (u *UseCase) GetBalance(ctx ctx.Context, userID string) (balance float64, err error) {
	balance, err = u.redisRepo.GetBalance(ctx.Context, userID)
	if err != nil {
		balance, err = u.reloadBalance(ctx.Context, userID)
	}

	return
}

func (u *UseCase) reloadBalance(ctx context.Context, userID string) (balance float64, err error) {
	balance, err = u.mysqlRepo.GetBalance(ctx, userID)
	if err != nil {
		return
	}

	noneReturnErr := u.redisRepo.SetBalance(ctx, userID, balance)
	if noneReturnErr != nil {
		zap.L().Error(noneReturnErr.Error())
	}

	return
}

func (u *UseCase) GetUserPointLogs(ctx ctx.Context, userID string) (result []domain.PointLog, err error) {
	result, err = u.redisRepo.GetUserPointLogs(ctx.Context, userID)
	if err != nil {
		result, err = u.reloadUserPointLogs(ctx.Context, userID)
	}

	return
}

func (u *UseCase) reloadUserPointLogs(ctx context.Context, userID string) (result []domain.PointLog, err error) {
	result, err = u.mysqlRepo.GetUserPointLogs(ctx, userID)
	if err != nil {
		return
	}

	noneReturnErr := u.redisRepo.SetUserPointLogs(ctx, userID, result)
	if noneReturnErr != nil {
		zap.L().Error(noneReturnErr.Error())
	}

	return
}

func (u *UseCase) UpdatePoint(ctx ctx.Context, userID, opCode, typ string, difference float64) (balance float64, err error) {
	tx, err := u.mysqlRepo.BeginTx(ctx.Context)
	if err != nil {
		return
	}

	total, err := u.mysqlRepo.GetUnmodifiedPoint(ctx.Context, tx, userID)
	if err != nil {
		return
	}
	// 判斷異動積分後不得小於 0
	if total+difference < 0 {
		err = catch.New(errno.InsufficientUserBalance, "insufficient User balance", fmt.Sprintf("insufficient User balance. balance: %f, difference: %f", balance, difference))
		return
	}
	err = u.mysqlRepo.UpdatePoint(ctx.Context, tx, userID, opCode, typ, total, difference)
	if err != nil {
		return
	}

	balance = total + difference
	err = u.mysqlRepo.Commit(tx)
	if err != nil {
		u.mysqlRepo.Rollback(tx)
	}

	noneReturnErr := u.redisRepo.SetBalance(ctx.Context, userID, balance)
	if noneReturnErr != nil {
		zap.L().Error(noneReturnErr.Error())
	}

	return
}
