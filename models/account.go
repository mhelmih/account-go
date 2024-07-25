package models

import (
	"gorm.io/gorm"
)

type Nasabah struct {
	gorm.Model
	NoRekening      string      `gorm:"primaryKey"`
	Nama            string      `gorm:"not null"`
	Nik             string      `gorm:"unique;not null"`
	NoHp            string      `gorm:"unique;not null"`
	Saldo           Saldo       `gorm:"foreignKey:NoRekening;references:NoRekening;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	DaftarTransaksi []Transaksi `gorm:"foreignKey:Id;references:NoRekening;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Saldo struct {
	gorm.Model
	NoRekening string  `gorm:"primaryKey"`
	Saldo      float64 `gorm:"not null"`
}

type Transaksi struct {
	gorm.Model
	Id            uint    `gorm:"primaryKey"`
	NoRekening    string  `gorm:"not null;index"`
	Nominal       float64 `gorm:"not null"`
	TipeTransaksi string  `gorm:"not null"`
}

type Counter struct {
	gorm.Model
	Name  string `gorm:"unique;not null"`
	Value int64  `gorm:"not null"`
}

type RegisterRequest struct {
	Nama string `json:"nama" validate:"required"`
	Nik  string `json:"nik" validate:"required"`
	NoHp string `json:"no_hp" validate:"required"`
}

type TrxRequest struct {
	NoRekening string  `json:"no_rekening" validate:"required"`
	Nominal    float64 `json:"nominal" validate:"required,gt=0"`
}
