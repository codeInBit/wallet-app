package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

//Wallet - Wallet struct that represents the Wallet model
type Wallet struct {
	gorm.Model
	UserID            uint `gorm:"not null" json:"user_id"`
	Balance           int  `gorm:"default:10000" json:"balance"`
	WalletTransaction []WalletTransaction
}

//FindUserWalletByID - Returns a user's walllet
func (w *Wallet) FindUserWalletByID(uid uint, db *gorm.DB) (*Wallet, error) {
	var err error
	wallet := Wallet{}

	err = db.Debug().Where("user_id = ?", uid).Take(&wallet).Error
	if err != nil {
		return &Wallet{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &Wallet{}, errors.New("Wallet Not Found")
	}
	return &wallet, err
}

//Transfer - Transfer Value between two users
func (w *Wallet) Transfer(senderID, recipientID uint, amount int, db *gorm.DB) error {
	var err error
	walletTransaction := WalletTransaction{}

	//Sender wallet
	senderWallet, err := w.FindUserWalletByID(senderID, db)
	if err != nil {
		return err
	}

	//Recipient Wallet
	recipientWallet, err := w.FindUserWalletByID(recipientID, db)
	if err != nil {
		return err
	}

	//Check if sender has enough value to transfer, if not, log the transaction as failed transaction
	if senderWallet.Balance/100 >= amount {
		//Debit recipient
		w.DebitWallet(senderWallet.ID, amount, db)

		//Credit Sender
		w.CreditWallet(recipientWallet.ID, amount, db)
	} else {
		//Log failed transaction based on insufficient balance
		err = walletTransaction.SaveTransaction(senderWallet.ID, amount, senderWallet.Balance, senderWallet.Balance, "non", "Cancelled", "Insufficient Balance", db)
		err = errors.New("Transfer cancelled insufficient balance")
		if err != nil {
			return err
		}
	}

	return nil
}

//CreditWallet - This method credits a user's wallet
func (w *Wallet) CreditWallet(walletID uint, amount int, db *gorm.DB) error {
	var err error
	walletTransaction := WalletTransaction{}

	//Get wallet
	wallet, err := w.FindUserWalletByID(walletID, db)
	if err != nil {
		return err
	}
	newAmount := amount * 100
	prevBalance := wallet.Balance
	currentBalance := wallet.Balance + newAmount

	// Update wallet balance
	db = db.Debug().Model(&Wallet{}).Where("id = ?", walletID).Take(&Wallet{}).UpdateColumns(
		map[string]interface{}{
			"balance": currentBalance,
		},
	)
	if db.Error != nil {
		return db.Error
	}

	//Log transaction
	err = walletTransaction.SaveTransaction(walletID, newAmount, prevBalance, currentBalance, "cr", "success", "Wallet credited", db)
	if err != nil {
		return err
	}

	return nil
}

//DebitWallet - This method debits a user's wallet
func (w *Wallet) DebitWallet(walletID uint, amount int, db *gorm.DB) error {
	var err error
	walletTransaction := WalletTransaction{}

	//Get wallet
	wallet, err := w.FindUserWalletByID(walletID, db)
	if err != nil {
		return err
	}
	newAmount := amount * 100
	prevBalance := wallet.Balance
	currentBalance := wallet.Balance - newAmount

	// Update wallet balance
	db = db.Debug().Model(&Wallet{}).Where("id = ?", walletID).Take(&Wallet{}).UpdateColumns(
		map[string]interface{}{
			"balance": currentBalance,
		},
	)
	if db.Error != nil {
		return db.Error
	}

	//Log transaction
	err = walletTransaction.SaveTransaction(walletID, newAmount, prevBalance, currentBalance, "cr", "success", "Wallet credited", db)
	if err != nil {
		return err
	}

	return nil
}
