package views

import (
	"path/filepath"
	"runtime"

	infra "github.com/SAIKAII/skResk-Infra"
	"github.com/SAIKAII/skResk-Infra/base"
	"github.com/SAIKAII/skResk-ui/core/users"
	"github.com/kataras/iris"
)

type View struct {
	UserService *users.UserService
	groupRouter iris.Party
}

func init() {
	infra.RegisterApi(&View{})
}

func (v *View) Init() {
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	app := base.Iris()
	app.StaticWeb("/public/static", filepath.Join(dir, "../static"))
	app.StaticWeb("/public/views", dir)
	v.groupRouter = app.Party("/v1/envelope")
	v.index()
	v.SendingRedEnvelopeIndex()
}

func (v *View) index() {
	base.Iris().Get("/index", func(ctx iris.Context) {
		ctx.View("ui/index.html")
	})
	base.Iris().Get("/home", func(ctx iris.Context) {
		ctx.View("ui/index.html")
	})
}

func (v *View) SendingRedEnvelopeIndex() {
	v.groupRouter.Get("/Sending", func(ctx iris.Context) {
		ctx.View("app.html")
	})
}
