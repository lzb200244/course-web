package service

import (
	"errors"
	"fmt"
	"go-template/global/auth"
	"go-template/global/code"
	"go-template/models"
	"go-template/models/response"
	"go-template/respository"
	"go-template/utils"
	"gorm.io/gorm"
	"sync"
)

/*
Created by 斑斑砖 on 2023/8/14.
Description：
	用户
*/

// ================================================================= 用户注册

type UserRegister struct {
	Username string
	Password string
	Email    string
}

func NewUser(username string, password string, email string) *UserRegister {
	return &UserRegister{
		Username: username, Password: password, Email: email,
	}
}
func Register(username string, password string, email string) (interface{}, code.Code) {
	return NewUser(username, password, email).Do()
}
func (r UserRegister) Do() (interface{}, code.Code) {
	_, c := r.checkExists()
	if c != code.OK {
		return nil, c
	}
	_, c = r.create()
	if c != code.OK {
		return nil, c
	}

	return nil, code.OK
}

// 用户是否存在
func (r UserRegister) checkExists() (interface{}, code.Code) {
	//TODO 邮箱为进行校验且存在unique
	_, err := respository.GetOne(&models.User{}, "user_name", r.Username)
	//不存在该用户
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.OK
		}
		//数据库异常错误
		return nil, code.ERROR_DB_OPE
	}
	return nil, code.ERROR_DB_OPE
}

// 创建用户
func (r UserRegister) create() (interface{}, code.Code) {
	user := models.User{UserName: r.Username, Password: r.Password, Email: r.Email}
	user.SetPassword() // 加密

	err := respository.Create(&user)

	//给用户赋予权限
	respository.AddUserAuthority(user, []int{int(auth.Student)})
	if err != nil {
		fmt.Println(err)
		//TODO 记录日志
		return nil, code.ERROR_DB_OPE
	}
	return nil, code.OK

}

// ================================================================= 用户登录

type UserLogin struct {
	Username string
	Password string
}

func NewUserLogin(username string, password string) *UserLogin {
	return &UserLogin{Username: username, Password: password}
}
func Login(username string, password string) (interface{}, code.Code) {
	return NewUserLogin(username, password).Do()
}
func (r UserLogin) Do() (interface{}, code.Code) {
	data, c := r.CheckAndSign()
	if c != code.OK {
		return nil, c
	}
	return data, code.OK
}
func (r UserLogin) CheckAndSign() (interface{}, code.Code) {

	userObj, err := respository.GetUserInfo(&models.User{}, "user_name", r.Username)
	if err != nil {
		//不存在该用户
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ERROR_USER_NOT_EXIST
		}
		//数据库异常错误
		return nil, code.ERROR_DB_OPE
	}
	//密码错误
	if ok := userObj.CheckPassword(r.Password); !ok {
		return nil, code.ERROR_PASSWORD_WRONG
	}
	//校验通过生成Token
	token, err := utils.GenerateToken(userObj.ID, userObj.UserName, userObj.Email, int(userObj.Roles[0].ID))
	if err != nil {
		//签发token失败
		return nil, code.ERROR_TOKEN_CREATE
	}
	roleObj := userObj.Roles[0]
	permission := respository.GetPermission(int(roleObj.ID))

	return response.NewUserResponse(
		userObj.ID,
		userObj.UserName,
		userObj.Name,
		userObj.Email,
		userObj.Desc,
		userObj.Avatar,
		token,
		userObj.Sex,
		[]string{roleObj.Name},
		permission,
	), code.OK

}

// ================================================================= 获取用户信息

type UserInfo struct {
}

func NewUserInfo() *UserInfo {
	return &UserInfo{}
}
func GetUserInfo(userID, roleID int) (interface{}, code.Code) {
	return NewUserInfo().Do(userID, roleID)
}
func (u *UserInfo) Do(userID, roleID int) (interface{}, code.Code) {
	//1. 判断用户是否存在
	userObj, c := u.GetUserObj(userID, roleID)
	if c != code.OK {
		return nil, c
	}

	return userObj, code.OK
}
func (u *UserInfo) GetUserObj(userID, roleID int) (interface{}, code.Code) {
	var wg sync.WaitGroup
	var permission []int
	var userObj *models.User
	var err error
	wg.Add(2)
	//1. 获取用户信息
	go func() {
		defer wg.Done()
		userObj, err = respository.GetUserInfo(&models.User{}, "id", userID)
	}()
	//2. 获取用户角色与权限
	go func() {
		defer wg.Done()
		permission = respository.GetPermission(roleID)
	}()
	wg.Wait()
	if err != nil {

		//不存在该用户
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ERROR_USER_NOT_EXIST
		}
		//数据库异常错误
		return nil, code.ERROR_DB_OPE
	}
	return response.NewUserResponse(
		userObj.ID,
		userObj.UserName,
		userObj.Name,
		userObj.Email,
		userObj.Desc,
		userObj.Avatar,
		"",
		userObj.Sex,
		[]string{auth.GetAuthorityName(roleID)},
		permission,
	), code.OK

}
