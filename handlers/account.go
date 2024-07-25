package handlers

import (
	"account/models"
	"fmt"
	"net/http"

	"log/slog"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Client for account service
type Client struct {
	Db     *gorm.DB
	Logger *slog.Logger
}

// NewClient creates a new client for account service
func NewClient(db *gorm.DB) *Client {
	return &Client{
		Db:     db,
		Logger: slog.Default(),
	}
}

// Register a person to the database
func (c *Client) Register(ctx echo.Context) error {
	req := new(models.RegisterRequest)
	if err := ctx.Bind(req); err != nil {
		c.Logger.Error("failed to bind the body request.", "err", err.Error())
		return ctx.JSON(http.StatusBadRequest, map[string]string{"remark": "failed to bind the body request", "err": err.Error()})
	}
	if err := ctx.Validate(req); err != nil {
		c.Logger.Error("failed to validate the body request.", "err", err.Error())
		return ctx.JSON(http.StatusBadRequest, map[string]string{"remark": "failed to validate the body request", "err": err.Error()})
	}

	var count int64
	err := c.Db.Model(&models.Nasabah{}).Where("nik = ? OR no_hp = ?", req.Nik, req.NoHp).Count(&count).Error
	if err != nil {
		c.Logger.Error("failed to check NIK or NoHp.", "err", err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"remark": "could not check NIK or NoHp", "err": err.Error()})
	}
	if count > 0 {
		c.Logger.Warn("NIK or NoHp is already used")
		return ctx.JSON(http.StatusBadRequest, map[string]string{"remark": "NIK or NoHp is already used"})
	}

	noRekening, err := c.generateNoRekening()
	if err != nil {
		c.Logger.Error("failed to generate account number.", "err", err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"remark": "failed to generate account number", "err": err.Error()})
	}

	nasabah := models.Nasabah{
		Nama:       req.Nama,
		Nik:        req.Nik,
		NoHp:       req.NoHp,
		NoRekening: noRekening,
		Saldo: models.Saldo{
			NoRekening: noRekening,
			Saldo:      0.0,
		},
	}
	if err := c.Db.Create(&nasabah).Error; err != nil {
		c.Logger.Error("failed to register nasabah", "err", err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"remark": "failed to register nasabah", "err": err.Error()})
	}

	c.Logger.Info("success to register nasabah with", "no_rekening", noRekening)
	return ctx.JSON(http.StatusOK, map[string]string{"no_rekening": noRekening})
}

// Deposit money to an account
func (c *Client) Deposit(ctx echo.Context) error {
	req := new(models.TrxRequest)
	if err := ctx.Bind(req); err != nil {
		c.Logger.Error("failed to bind the body request.", "err", err.Error())
		return ctx.JSON(http.StatusBadRequest, map[string]string{"remark": "failed to bind the body request", "err": err.Error()})
	}
	if err := ctx.Validate(req); err != nil {
		c.Logger.Error("failed to validate the body request.", "err", err.Error())
		return ctx.JSON(http.StatusBadRequest, map[string]string{"remark": "failed to validate the body request", "err": err.Error()})
	}

	saldoNasabah := models.Saldo{}
	tx := c.Db.Begin()
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("no_rekening = ?", req.NoRekening).Find(&saldoNasabah).Error; err != nil {
		tx.Rollback()
		c.Logger.Error(fmt.Sprintf("failed to get nasabah with no_rekening=%s. err=%s", req.NoRekening, err.Error()))
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"remark": fmt.Sprintf("failed to get nasabah with no_rekening=%s", req.NoRekening), "err": err.Error()})
	}
	if saldoNasabah.NoRekening == "" {
		tx.Rollback()
		c.Logger.Warn(fmt.Sprintf("failed to find nasabah with no_rekening=%s (not found)", req.NoRekening))
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{"remark": fmt.Sprintf("failed to find nasabah with no_rekening=%s (not found)", req.NoRekening)})
	}

	if err := tx.Model(&saldoNasabah).Update("saldo", gorm.Expr("saldo + ?", req.Nominal)).Error; err != nil {
		tx.Rollback()
		c.Logger.Error(fmt.Sprintf("failed to deposit nominal=%f to nasabah with no_rekening=%s. err=%s", req.Nominal, req.NoRekening, err.Error()))
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"remark": fmt.Sprintf("failed to deposit nominal=%f to nasabah with no_rekening=%s", req.Nominal, req.NoRekening), "err": err.Error()})
	}

	transaksi := models.Transaksi{
		NoRekening:    req.NoRekening,
		Nominal:       req.Nominal,
		TipeTransaksi: "D",
	}
	if err := tx.Create(&transaksi).Error; err != nil {
		c.Logger.Error("failed to record the deposit transaction. ", "err", err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"remark": "failed to record the transaction", "err": err.Error()})
	}

	if err := tx.Commit().Error; err != nil {
		c.Logger.Error("failed to commit transaction", ". err: ", err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"remark": "failed to commit transaction", "err": err.Error()})
	}

	c.Logger.Info(fmt.Sprintf("success to deposit nominal=%f to nasabah with no_rekening=%s", req.Nominal, req.NoRekening))
	return ctx.JSON(http.StatusOK, map[string]interface{}{"saldo": saldoNasabah.Saldo + req.Nominal})
}

// Withdraw money from an account
func (c *Client) Withdraw(ctx echo.Context) error {
	req := new(models.TrxRequest)
	if err := ctx.Bind(req); err != nil {
		c.Logger.Error("failed to bind the body request.", "err", err.Error())
		return ctx.JSON(http.StatusBadRequest, map[string]string{"remark": "failed to bind the body request", "err": err.Error()})
	}
	if err := ctx.Validate(req); err != nil {
		c.Logger.Error("failed to validate the body request.", "err", err.Error())
		return ctx.JSON(http.StatusBadRequest, map[string]string{"remark": "failed to validate the body request", "err": err.Error()})
	}

	saldoNasabah := models.Saldo{}
	tx := c.Db.Begin()
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("no_rekening = ?", req.NoRekening).Find(&saldoNasabah).Error; err != nil {
		tx.Rollback()
		c.Logger.Error(fmt.Sprintf("failed to get nasabah with no_rekening=%s. err=%s", req.NoRekening, err.Error()))
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"remark": fmt.Sprintf("failed to get nasabah with no_rekening=%s", req.NoRekening), "err": err.Error()})
	}
	if saldoNasabah.NoRekening == "" {
		tx.Rollback()
		c.Logger.Warn(fmt.Sprintf("failed to find nasabah with no_rekening=%s (not found)", req.NoRekening))
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{"remark": fmt.Sprintf("failed to find nasabah with no_rekening=%s (not found)", req.NoRekening)})
	}
	if saldoNasabah.Saldo < req.Nominal {
		tx.Rollback()
		c.Logger.Warn(fmt.Sprintf("failed to withdraw money with nominal=%f from account with smaller balance saldo=%f", req.Nominal, saldoNasabah.Saldo))
	}

	if err := tx.Model(&saldoNasabah).Update("saldo", gorm.Expr("saldo - ?", req.Nominal)).Error; err != nil {
		tx.Rollback()
		c.Logger.Error(fmt.Sprintf("failed to withdraw nominal=%f from nasabah with no_rekening=%s. err=%s", req.Nominal, req.NoRekening, err.Error()))
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"remark": fmt.Sprintf("failed to withdraw nominal=%f from nasabah with no_rekening=%s", req.Nominal, req.NoRekening), "err": err.Error()})
	}

	transaksi := models.Transaksi{
		NoRekening:    req.NoRekening,
		Nominal:       req.Nominal,
		TipeTransaksi: "C",
	}
	if err := tx.Create(&transaksi).Error; err != nil {
		tx.Rollback()
		c.Logger.Error("failed to record the withdraw transaction. ", "err", err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"remark": "failed to record the transaction", "err": err.Error()})
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.Logger.Error("failed to commit transaction", ". err: ", err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"remark": "failed to commit transaction", "err": err.Error()})
	}

	c.Logger.Info(fmt.Sprintf("success to withdraw nominal=%f to nasabah with no_rekening=%s", req.Nominal, req.NoRekening))
	return ctx.JSON(http.StatusOK, map[string]interface{}{"saldo": saldoNasabah.Saldo - req.Nominal})
}

// Check balance of an account
func (c *Client) CheckBalance(ctx echo.Context) error {
	noRekening := ctx.Param("no_rekening")
	saldoNasabah := models.Saldo{}

	err := c.Db.First(&saldoNasabah, "no_rekening = ?", noRekening)
	if err != nil {
		c.Logger.Warn(fmt.Sprintf("failed to find nasabah with no_rekening=%s (not found)", noRekening))
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{"remark": fmt.Sprintf("failed to find nasabah with no_rekening=%s (not found)", noRekening)})
	}

	c.Logger.Info(fmt.Sprintf("success to get account balance with no_rekening=%s", noRekening))
	return ctx.JSON(http.StatusOK, map[string]interface{}{"saldo": saldoNasabah.Saldo})
}

// Generate 10 digits auto-increment account number
func (c *Client) generateNoRekening() (string, error) {
	var counter models.Counter
	tx := c.Db.Begin()
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("name = ?", "rekening").FirstOrCreate(&counter, models.Counter{Name: "rekening"}).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	counter.Value++
	if err := tx.Save(&counter).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", err
	}

	return fmt.Sprintf("%010d", counter.Value), nil
}
