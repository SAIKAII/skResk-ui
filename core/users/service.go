package users

import (
	"strconv"

	"github.com/SAIKAII/skResk-Account/services"
	"github.com/SAIKAII/skResk-Infra/base"
	"github.com/sirupsen/logrus"
	"github.com/tietang/dbx"
)

type UserService struct {
}

func (u *UserService) Login(mobile, username string) (user *User) {
	as := services.GetAccountService()
	err := base.Tx(func(runner *dbx.TxRunner) error {
		dao := UserDao{runner: runner}
		user = dao.GetOne(mobile)
		// 创建用户
		if user == nil {
			user = &User{
				Mobile:   mobile,
				Username: username,
			}
			id, err := dao.Insert(user)
			if err != nil {
				logrus.Error(err)
				return err
			}
			user.Id = id
			user.UserId = strconv.Itoa(int(id))
		}
		if user.Username != username {
			user.Username = username
			rows, err := dao.Update(user)
			if err != nil {
				logrus.Error(err)
				return err
			}
			if rows <= 0 {
				logrus.Warn("non updated: ", user)
			}
		}
		return nil
	})
	if err != nil {
		logrus.Error(err)
		return nil
	}
	user.UserId = strconv.Itoa(int(user.Id))
	// 创建资金账户
	a := as.GetEnvelopeAccountByUserId(user.UserId)
	if a == nil {
		dto := services.AccountCreatedDTO{
			UserId:       user.UserId,
			Username:     user.Username,
			AccountName:  user.Username,
			AccountType:  int(services.EnvelopeAccountType),
			CurrencyCode: "CNY",
			Amount:       "1000",
		}
		_, err := as.CreateAccount(dto)
		if err != nil {
			logrus.Error(err)
			return
		}
	} else {
		logrus.Info("已经存在红包资金账户: ", user)
	}
	return
}
