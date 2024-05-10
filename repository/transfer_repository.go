package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/yafireyhan01/e-wallet/model"
	"github.com/yafireyhan01/e-wallet/model/dto"
)

type TransferRepository interface {
	Create(payload dto.TransferRequest, send, receive model.User) (model.Transfer, error)
	GetSend(id string, page int) ([]model.Transfer, error)
	GetReceive(id string, page int) ([]model.Transfer, error)
	CreateWithdraw(payload model.Withdraw, saldo model.UserSaldo) (model.Withdraw, error)
	GetWithdraw(id string, page int) ([]model.Withdraw, error)
}

type transferRepository struct {
	db *sql.DB
}

// tulis code kalian disini
func (t *transferRepository) Create(payload dto.TransferRequest, send, receive model.User) (model.Transfer, error) {
	response := model.Transfer{}
	tx, err := t.db.Begin()
	if err != nil {
		tx.Rollback()
		return model.Transfer{}, err
	}
	send.Saldo -= payload.JumlahTransfer
	receive.Saldo += payload.JumlahTransfer
	if send.Saldo < 0 {
		tx.Rollback()
		return model.Transfer{}, fmt.Errorf("saldo anda tidak mendcukupi untuk transfer %d", payload.JumlahTransfer)
	}

	// buat catatan penerima ke database
	err = tx.QueryRow(`INSERT INTO trx_send_transfer (
		user_id,
		tujuan_transfer,
		jumlah_transfer,
		jenis_transfer,
		transfer_at
		)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
	) RETURNING id`, send.Id, receive.Id, payload.JumlahTransfer, "mengirim", time.Now()).Scan(&response.Id)
	if err != nil {
		tx.Rollback()
		return model.Transfer{}, err
	}

	_, err = tx.Exec(`UPDATE mst_saldo SET saldo=$1 WHERE user_id=$2`, receive.Saldo, receive.Id)
	if err != nil {
		tx.Rollback()
		return model.Transfer{}, err
	}
	_, err = tx.Exec(`UPDATE mst_saldo SET saldo = $1 WHERE user_id=$2`, send.Saldo, send.Id)
	if err != nil {
		tx.Rollback()
		return model.Transfer{}, err
	}
	response.JenisTransfer = "mengirim"
	_, err = tx.Query(`INSERT INTO trx_receive_transfer (
		user_id,
		trx_id,
		tujuan_transfer,
		jumlah_transfer,
		jenis_transfer,
		transfer_at)
	VALUES ($1,$2,$3,$4,$5,$6)`, send.Id, response.Id, receive.Id, payload.JumlahTransfer, "menerima", time.Now())
	if err != nil {
		tx.Rollback()
		return model.Transfer{}, err
	}

	response.UserId = send.Id
	response.TujuanTransfer = receive.Id
	response.JumlahTransfer = payload.JumlahTransfer
	tx.Commit()

	return response, nil
}

func (t *transferRepository) GetSend(id string, page int) ([]model.Transfer, error) {
	var datas []model.Transfer
	paging := 3
	limit := (paging * page) - paging

	res, err := t.db.Query(`SELECT 
		trx.id,
		trx.user_id,
		mst_user.name,
		trx.tujuan_transfer,
		mst_tujuan.name,
		trx.jumlah_transfer,
		trx.jenis_transfer
	FROM 
		trx_send_transfer AS trx
	LEFT JOIN 
		mst_user ON trx.user_id = mst_user.id
	LEFT JOIN 
		mst_user AS mst_tujuan ON trx.tujuan_transfer = mst_tujuan.id
	WHERE 
		trx.user_id = $1
	ORDER BY 
		trx.transfer_at DESC 
	LIMIT $2 OFFSET $3`, id, paging, limit)
	if err != nil {
		return []model.Transfer{}, err
	}
	defer res.Close()

	for res.Next() {
		var data model.Transfer
		err := res.Scan(&data.Id, &data.UserId, &data.SenderName, &data.TujuanTransfer, &data.Receiver, &data.JumlahTransfer, &data.JenisTransfer)
		if err != nil {
			return []model.Transfer{}, err
		}
		datas = append(datas, data)
	}

	return datas, nil
}

func (t *transferRepository) GetReceive(id string, page int) ([]model.Transfer, error) {
	var datas []model.Transfer
	paging := 3
	limit := (paging * page) - paging

	res, err := t.db.Query(`SELECT 
		trx.id,
		trx.user_id,
		mst_user.name,
		trx.trx_id,
		trx.tujuan_transfer,
		mst_tujuan.name,
		trx.jumlah_transfer,
		trx.jenis_transfer
	FROM 
    	trx_receive_transfer AS trx
	LEFT JOIN 
    	mst_user ON trx.user_id = mst_user.id
	LEFT JOIN 
    	mst_user AS mst_tujuan ON trx.tujuan_transfer = mst_tujuan.id
	WHERE 
    	trx.tujuan_transfer = $1
	ORDER BY 
    	trx.transfer_at DESC 
	LIMIT $2 OFFSET $3`, id, paging, limit)
	if err != nil {
		return []model.Transfer{}, err
	}
	defer res.Close()

	for res.Next() {
		var data model.Transfer
		err := res.Scan(&data.Id, &data.UserId, &data.SenderName, &data.Trx_id, &data.TujuanTransfer, &data.Receiver, &data.JumlahTransfer, &data.JenisTransfer)
		if err != nil {
			return []model.Transfer{}, err
		}
		datas = append(datas, data)
	}

	return datas, nil
}

func (t *transferRepository) CreateWithdraw(payload model.Withdraw, saldo model.UserSaldo) (model.Withdraw, error) {
	response := model.Withdraw{}
	tx, _ := t.db.Begin()
	err := tx.QueryRow(`INSERT INTO withdraw_saldo (user_id,withdraw,created_at)
	VALUES
		($1,$2,$3)
	RETURNING 
		id,user_id,withdraw,created_at`, payload.UserId, payload.Withdraw, time.Now()).Scan(
		&response.Id,
		&response.UserId,
		&response.Withdraw,
		&response.Created_at,
	)
	if err != nil {
		tx.Rollback()
		return model.Withdraw{}, err
	}
	_, err = tx.Exec(`UPDATE mst_saldo SET saldo=$1 WHERE user_id=$2`, saldo.Saldo, payload.UserId)
	if err != nil {
		tx.Rollback()
		return model.Withdraw{}, err
	}

	tx.Commit()

	return response, nil
}

func (t *transferRepository) GetWithdraw(id string, page int) ([]model.Withdraw, error) {
	var datas []model.Withdraw

	paging := 3
	limit := (paging * page) - paging

	res, err := t.db.Query(`SELECT 
		id,
		user_id,
		withdraw,
		created_at
	FROM 
    	withdraw_saldo
	WHERE 
    	user_id = $1
	ORDER BY 
    	created_at DESC 
	LIMIT $2 OFFSET $3`, id, paging, limit)
	if err != nil {
		return []model.Withdraw{}, err
	}
	defer res.Close()

	for res.Next() {
		var data model.Withdraw

		err := res.Scan(&data.Id, &data.UserId, &data.Withdraw, &data.Created_at)
		if err != nil {
			return []model.Withdraw{}, err
		}
		datas = append(datas, data)
	}

	return datas, nil
}

func NewTransferRepository(db *sql.DB) TransferRepository {
	return &transferRepository{db: db}
}
