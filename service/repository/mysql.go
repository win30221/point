package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/win30221/code/errno"
	"github.com/win30221/core/http/catch"
	"github.com/win30221/core/syserrno"
	"github.com/win30221/point/domain"
)

type MysqlRepo struct {
	master *sql.DB
	slave  *sql.DB
}

func NewMysqlRepo(master, slave *sql.DB) *MysqlRepo {
	return &MysqlRepo{master, slave}
}

func (r *MysqlRepo) BeginTx(ctx context.Context) (tx *sql.Tx, err error) {
	tx, err = r.master.BeginTx(ctx, nil)
	if err != nil {
		err = catch.New(syserrno.MySQL, "begin tx error", err.Error())
		return
	}
	return
}

func (r *MysqlRepo) Commit(tx *sql.Tx) (err error) {
	err = tx.Commit()
	if err != nil {
		err = catch.New(syserrno.MySQL, "commit tx error", err.Error())
		return
	}

	return
}

func (r *MysqlRepo) Rollback(tx *sql.Tx) {
	tx.Rollback()
}

// 建立積分錢包
func (r *MysqlRepo) CreateWallet(ctx context.Context, data domain.CreateWalletReq) (err error) {
	var values []interface{}

	query := "INSERT INTO point (`userID`) VALUES (?)" + strings.Repeat(", (?)", len(data.UserID)-1) + ";"
	for _, v := range data.UserID {
		values = append(values, v)
	}

	_, err = r.master.ExecContext(ctx, query, values...)

	mysqlError, ok := err.(*mysql.MySQLError)
	if ok && mysqlError.Number == 1062 {
		err = catch.New(errno.ExistedPointWallet, "duplicated point wallet", err.Error())

		return
	}
	if err != nil {
		err = catch.New(syserrno.MySQL, "create point wallet error", err.Error())

		return
	}

	return
}

// 取得積分
func (r *MysqlRepo) GetBalance(ctx context.Context, userID string) (balance float64, err error) {
	row := r.master.QueryRowContext(ctx, "SELECT `balance` FROM point WHERE `userID` = ?;", userID)
	err = row.Scan(&balance)
	if err != nil {
		err = catch.New(syserrno.MySQL, "get balance error", err.Error())
		return
	}
	return
}

// 取得未修改積分
func (r *MysqlRepo) GetUnmodifiedPoint(ctx context.Context, tx *sql.Tx, userID string) (balance float64, err error) {
	row := tx.QueryRowContext(ctx, "SELECT `balance` FROM point WHERE `userID` = ? FOR UPDATE;", userID)
	err = row.Scan(&balance)
	if err != nil {
		err = catch.New(syserrno.MySQL, "get point error", err.Error())
		return
	}
	return
}

// 異動積分
func (r *MysqlRepo) UpdatePoint(ctx context.Context, tx *sql.Tx, userID, opCode, typ string, before, difference float64) (err error) {
	// 增加積分異動 log
	_, err = tx.ExecContext(ctx, "INSERT INTO pointLog(`userID`, `opCode`, `type`, `before`, `difference`) values (?, ?, ?, ?, ?);", userID, opCode, typ, before, difference)
	if err != nil {
		err = catch.New(syserrno.MySQL, "change point error", err.Error())
		return
	}
	// 異動積分
	_, err = tx.ExecContext(ctx, "UPDATE point SET `balance` = `balance` + ? WHERE `userID` = ?;", difference, userID)
	if err != nil {
		err = catch.New(syserrno.MySQL, "change point error", err.Error())
		return
	}

	return
}

// 取得積分異動紀錄
func (r *MysqlRepo) GetUserPointLogs(ctx context.Context, userID string) (result []domain.PointLog, err error) {
	rows, err := r.slave.QueryContext(ctx, "SELECT userID, opCode, `type`, `before`, difference, createdAt FROM pointLog WHERE userID = ? ORDER BY id DESC LIMIT 20;", userID)
	if err != nil {
		err = catch.New(syserrno.MySQL, "get User point logs error", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		res := domain.PointLog{}
		err = rows.Scan(
			&res.UserID,
			&res.OpCode,
			&res.Type,
			&res.Before,
			&res.Difference,
			&res.CreatedAt,
		)
		if err != nil {
			err = catch.New(syserrno.MySQL, "get User point logs error", err.Error())
			return
		}
		result = append(result, res)
	}
	return
}
