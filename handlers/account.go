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
		return ctx.JSON(http.StatusBadRequest, map[string]string{"remark": "could not check NIK or NoHp", "err": err.Error()})
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
		Saldo:      0.0,
	}

	if err := c.Db.Create(&nasabah).Error; err != nil {
		c.Logger.Error("failed to register nasabah", "err", err.Error())
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"remark": "failed to register nasabah", "err": err.Error()})
	}

	c.Logger.Info("success to register nasabah with", "no_rekening", noRekening)
	return ctx.JSON(http.StatusOK, map[string]string{"no_rekening": noRekening})
}

func (c *Client) Deposit(ctx echo.Context) error {
	// Implementasi logika menabung
	return ctx.JSON(http.StatusOK, map[string]interface{}{"saldo": 1000})
}

func (c *Client) Withdraw(ctx echo.Context) error {
	// Implementasi logika penarikan
	return ctx.JSON(http.StatusOK, map[string]interface{}{"saldo": 500})
}

func (c *Client) CheckBalance(ctx echo.Context) error {
	// Implementasi logika cek saldo
	return ctx.JSON(http.StatusOK, map[string]interface{}{"saldo": 500})
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

	tx.Commit()
	return fmt.Sprintf("%010d", counter.Value), nil
}
