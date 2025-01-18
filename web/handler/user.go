package handler

import (
	"ht/model"
	"ht/server"
	"ht/web/view/screens"
	"time"

	"github.com/labstack/echo/v4"
)

type UserView struct {
	server *server.Server
}

func NewUserView(server *server.Server) *UserView {
	newUserView := &UserView{
		server: server,
	}
	return newUserView
}

func (r *UserView) HandleUser(c echo.Context) error {
	return render(c, screens.User(&model.User{CreatedAt: time.Now()}))
}

// func (r *UserView) HandleUserBody(c echo.Context) error {
// 	auths, err := r.server.AuthService.GetAuths(c)
// 	if err != nil {
// 		return echo.NewHTTPError(http.StatusNotFound, err)
// 	}

// 	var tableDatas []model.Mapper
// 	for _, v := range auths {
// 		tableDatas = append(tableDatas, v)
// 	}
// 	var body = screens.BodyUser(model.DatamodelForUser, tableDatas)

// 	return render(c, body)
// }

// func (r *UserView) HandlePopupCreateUser(c echo.Context) error {
// 	projectRid := helper.GetRequestContext(c.Request().Context()).ProjectRID

// 	return renderPopup(c, components.PopupCreate(model.DatamodelForUser, "/project/"+projectRid.String()+"/user/createUser"))
// }
