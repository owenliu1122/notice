package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/owenliu1122/notice"

	"github.com/fpay/foundation-go/log"
	"github.com/labstack/echo"
)

// NewUserController returns an user table operation controller.
func NewUserController(logger *log.Logger, us notice.UserServiceInterface) *UserController {
	return &UserController{
		logger: logger,
		svc:    us,
	}
}

// UserController is an user table operation controller.
type UserController struct {
	logger *log.Logger
	svc    notice.UserServiceInterface
}

// List will return all users in the users table.
func (ctl *UserController) List(ctx echo.Context) error {
	userName := ctx.QueryParam("name")

	pageStr := ctx.QueryParam("page")
	page, e := strconv.Atoi(pageStr)
	if e != nil {
		ctl.logger.Errorf("page string param convert to int, page: %s, err: %s", pageStr, e)
		return ctx.String(http.StatusBadRequest, e.Error())
	}

	pageSizeStr := ctx.QueryParam("page_size")
	pageSize, e := strconv.Atoi(pageSizeStr)
	if e != nil {
		ctl.logger.Errorf("page size string param convert to int, page size: %s err: %s", pageSizeStr, e)
		return ctx.String(http.StatusBadRequest, e.Error())
	}

	res, cnt, err := ctl.svc.List(userName, page, pageSize)

	if err != nil {
		ctl.logger.Error("get users list failed, err: ", err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	ctl.logger.Debugf("user list, page: %d, pagesize: %d, cnt: %d\n", page, pageSize, cnt)
	ctl.logger.Debugf("user list, res: %v\n", res)

	return ctx.JSON(http.StatusOK, map[string]interface{}{"count": cnt, "data": res})
}

// Create will insert a new user record.
func (ctl *UserController) Create(ctx echo.Context) error {

	var user notice.User
	if err := ctx.Bind(&user); err != nil {
		ctl.logger.Error("add user get body failed, err: ", err)
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	ctl.logger.Infof("UserController Bind -> user: %v\n", user)

	if user.Name == "" {
		err := errors.New("create user failed, err: no user name")
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	if user.Email == "" && user.Phone == "" && user.Wechat == "" {
		err := errors.New("create user failed, did not fill in any communication method")
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	err := ctl.svc.Create(&user)

	if err != nil {
		ctl.logger.Error("create user failed, err: ", err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, user)
}

// Update will update an user record.
func (ctl *UserController) Update(ctx echo.Context) error {

	var user notice.User
	if err := ctx.Bind(&user); err != nil {
		ctl.logger.Error("update group get body failed, err: ", err)
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	ctl.logger.Infof("GroupController Update -> user: %#v\n", user)

	if user.ID == 0 {
		err := errors.New("update user failed, err: no user id")
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	if user.Name == "" && user.Email == "" && user.Phone == "" && user.Wechat == "" {
		err := errors.New("update user failed, did not fill in any modification information")
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	err := ctl.svc.Update(&user)

	if err != nil {
		ctl.logger.Error("update user failed, err: ", err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, user)
}

// Delete will delete an user record.
func (ctl *UserController) Delete(ctx echo.Context) error {

	var user notice.User
	if err := ctx.Bind(&user); err != nil {
		ctl.logger.Error("delete user get body failed, err: ", err)
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	ctl.logger.Infof("GroupController Delete -> user: %#v\n", user)

	if user.ID == 0 {
		err := errors.New("deleted user failed, err: no user id")
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	err := ctl.svc.Delete(&user)

	if err != nil {
		ctl.logger.Error("delete user failed, err: ", err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, user)
}
