package controller

import (
	"github.com/gin-gonic/gin"
	"go-template/global/code"
	"go-template/models/request"
	"go-template/service"
	"go-template/utils"
)

/*
Created by 斑斑砖 on 2023/8/14.
Description：
	注册
*/

func RegisterController(ctx *gin.Context) {
	//参数校验
	validate, err := utils.BindValidJson[request.Register](ctx)
	//参数校验失败
	if err != nil {
		utils.Fail(ctx, code.ERROR_REQUEST_PARAM, err.Error(), nil)
		return
	}
	_, c := service.Register(validate.Username, validate.Password, validate.Email)
	if c != code.OK {
		utils.Fail(ctx, c, err.Error(), nil)
		return
	}
	utils.Success(ctx, code.GetMsg(c), nil)
}
func LoginController(ctx *gin.Context) {
	//1. 参数校验
	validate, err := utils.BindValidJson[request.Login](ctx)
	//参数校验失败
	if err != nil {
		utils.Fail(ctx, code.ERROR_REQUEST_PARAM, err.Error(), nil)
		return
	}
	//2. 调用服务
	data, c := service.Login(validate.Username, validate.Password)
	if c != code.OK {
		utils.Fail(ctx, c, code.GetMsg(c), nil)
		return
	}
	utils.Success(ctx, code.GetMsg(c), data)
}

func GetUserController(ctx *gin.Context) {
	//1. 从ctx提取用户id
	user, _ := utils.GetUser(ctx)
	//2. 无需进行校验， 调用服务
	data, c := service.GetUserInfo(int(user.ID), user.Authority)
	if c != code.OK {
		utils.Fail(ctx, c, code.GetMsg(c), nil)
		return
	}
	utils.Success(ctx, code.GetMsg(c), data)
}