package views

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"github.com/SAIKAII/skResk-Envelope/services"

	"github.com/sirupsen/logrus"

	infra "github.com/SAIKAII/skResk-Infra"
	"github.com/SAIKAII/skResk-Infra/base"
	"github.com/SAIKAII/skResk-ui/core/users"
	"github.com/kataras/iris"
)

type MobileView struct {
	UserService *users.UserService
	groupRouter iris.Party
}

func init() {
	infra.RegisterApi(&MobileView{})
}

func (m *MobileView) Init() {
	m.UserService = &users.UserService{}
	dir := filepath.Join("./", "public/ui")
	app := base.Iris()
	//views := iris.HTML(dir, ".html")
	//// 开启开发者模式，每次请求都重新加载templates
	//views.Reload(true)
	//app.RegisterView(views)
	app.Favicon(filepath.Join(dir, "favicon.ico"))
	app.StaticWeb("/public/static", filepath.Join(dir, "../static"))
	app.StaticWeb("/public/ui", dir)
	m.groupRouter = app.Party("/envelope")
	m.groupRouter.Use(func(ctx iris.Context) {
		userId := ctx.GetCookie("userId")
		if userId == "" {
			ctx.Redirect("/login")
		} else {
			ctx.Next()
		}
	})
	// 登入登出
	app.Get("/login", m.loginHandler)
	app.Post("/login", m.loginSubmitHandler)
	app.Get("/logout", m.logoutHandler)
	// 我发的红包列表
	m.groupRouter.Get("/home", m.homeHandler)
	// 我抢到的红包列表
	m.groupRouter.Get("/recvd/list", m.receivedListHandler)
	// 红包记录
	m.groupRouter.Get("/list", m.listHandler)
	// 红包详情
	m.groupRouter.Get("/details", m.detailsHandler)
	// 可抢红包
	m.groupRouter.Get("/rev/home", m.receiveHomeHandler)
	// 抢红包
	m.groupRouter.Get("/recd", m.receiveSubmitHandler)
	// 发红包
	m.groupRouter.Get("/sending", m.sendingHandler)
	m.groupRouter.Post("/sending", m.sendingSubmitHandler)
}

func (m *MobileView) logoutHandler(ctx iris.Context) {
	ctx.RemoveCookie("userId")
	ctx.RemoveCookie("username")

	ctx.View("ui/index.html")
}

func (m *MobileView) loginHandler(ctx iris.Context) {
	ctx.View("ui/index.html")
}

func (m *MobileView) loginSubmitHandler(ctx iris.Context) {
	form := UserForm{}
	err := ctx.ReadForm(&form)
	if err != nil {
		logrus.Error(err)
	}
	if form.Mobile == "" {
		ctx.ViewData("msg", "手机号码不能为空！")
		ctx.View("index.html")
		return
	}
	if form.Username == "" {
		ctx.ViewData("msg", "用户名称不能为空！")
		ctx.View("index.html")
		return
	}
	user := m.UserService.Login(form.Mobile, form.Username)
	if user == nil {
		ctx.ViewData("msg", "系统出错了！")
		ctx.View("index.html")
		logrus.Info(user)
		return
	}

	ctx.SetCookieKV("userId", user.UserId, iris.CookieExpires(1*time.Hour))
	ctx.SetCookieKV("username", user.Username, iris.CookieExpires(1*time.Hour))

	ctx.Redirect("/envelope/home")
}

func (m *MobileView) homeHandler(ctx iris.Context) {
	userId := ctx.GetCookie("userId")
	es := services.GetRedEnvelopeService()
	orders := es.ListSent(userId, 0, 200)
	ctx.ViewData("orders", orders)
	ctx.ViewData("format", services.DefaultTimeFarmat)
	_ = ctx.View("ui/home.html")
}

// 我抢到的红包列表：recv_list.html /recvd/list
func (m *MobileView) receivedListHandler(ctx iris.Context) {
	userId := ctx.GetCookie("userId")
	es := services.GetRedEnvelopeService()
	items := es.ListReceived(userId, 0, 100)
	ctx.ViewData("items", items)
	ctx.ViewData("format", services.DefaultTimeFarmat)

	ctx.View("ui/recvd_list.html")
}

// 红包记录：re_one.html /list
func (m *MobileView) listHandler(ctx iris.Context) {
	envelopeNo := ctx.URLParamTrim("id")
	es := services.GetRedEnvelopeService()
	order := es.Get(envelopeNo)
	if order != nil {
		items := es.ListItems(envelopeNo)
		totalAmount := decimal.NewFromFloat(0)
		t1, t2 := time.Unix(int64(0), int64(0)), time.Unix(int64(0), int64(0))
		for i, v := range items {
			if i == 0 {
				t1 = v.CreatedAt
				t2 = v.CreatedAt
			} else {
				if t1.After(v.CreatedAt) {
					t1 = v.CreatedAt
				}
				if t2.Before(v.CreatedAt) {
					t2 = v.CreatedAt
				}
			}
			totalAmount = totalAmount.Add(v.Amount)
		}
		ctx.ViewData("items", items)
		ctx.ViewData("size", len(items))
		ctx.ViewData("totalAmount", totalAmount)
		seconds := t2.UnixNano() - t1.UnixNano()
		h := seconds / int64(time.Hour)
		msg := ""
		if h > 0 {
			msg += strconv.Itoa(int(h)) + "小时"
			seconds -= h * int64(time.Hour)
		}
		m := seconds / int64(time.Minute)

		if m > 0 {
			msg += strconv.Itoa(int(m)) + "分钟"
			seconds -= m * int64(time.Minute)
		}
		s := seconds / int64(time.Second)
		if s > 0 {
			msg += strconv.Itoa(int(s)) + "秒"
		}
		if msg == "" {
			msg = "0秒"
		}
		fmt.Println(t1, t2, seconds)
		ctx.ViewData("timeMsg", msg)
		ctx.ViewData("isReceived", len(items) == order.Quantity)
		ctx.ViewData("remainQuantity", order.Quantity-len(items))
	}
	ctx.ViewData("order", order)
	ctx.ViewData("hasOrder", order != nil)
	ctx.ViewData("format", services.DefaultTimeFarmat)
	ctx.View("ui/re_one.html")
}

// 红包详情：re_details.html /details
func (m *MobileView) detailsHandler(ctx iris.Context) {
	envelopeNo := ctx.URLParamTrim("id")
	es := services.GetRedEnvelopeService()
	order := es.Get(envelopeNo)
	ctx.ViewData("order", order)
	ctx.ViewData("hasOrder", order != nil)
	ctx.ViewData("format", services.DefaultTimeFarmat)
	ctx.View("ui/re_details.html")
}

// 可抢红包：rev_home.html /rev/home
func (m *MobileView) receiveHomeHandler(ctx iris.Context) {
	es := services.GetRedEnvelopeService()
	orders := es.ListReceivable(0, 200)
	ctx.ViewData("orders", orders)
	ctx.ViewData("hasOrder", len(orders) > 0)
	ctx.ViewData("format", services.DefaultTimeFarmat)
	ctx.View("ui/rev_home.html")
}

// 抢红包
func (m *MobileView) receiveSubmitHandler(ctx iris.Context) {
	envelopeNo := ctx.URLParamTrim("id")
	userId := ctx.GetCookie("userId")
	username := ctx.GetCookie("username")

	es := services.GetRedEnvelopeService()
	dto := services.RedEnvelopeReceiveDTO{
		EnvelopeNo:   envelopeNo,
		RecvUsername: username,
		RecvUserId:   userId,
	}
	item, err := es.Receive(dto)
	msg := ""
	if err == nil {
		ctx.ViewData("hasReceived", true)
	} else {
		ctx.ViewData("hasReceived", false)
		msg = err.Error()
	}
	order := es.Get(envelopeNo)
	ctx.ViewData("order", order)
	ctx.ViewData("item", item)
	ctx.ViewData("hasOrder", order != nil)
	ctx.ViewData("format", services.DefaultTimeFarmat)
	ctx.ViewData("msg", msg)
	ctx.View("ui/recd.html")
}

// 发红包
func (m *MobileView) sendingHandler(ctx iris.Context) {
	ctx.View("ui/sending.html")
}

// 发红包
func (m *MobileView) sendingSubmitHandler(ctx iris.Context) {
	form := RedEnvelopeSendingForm{}
	err := ctx.ReadForm(&form)
	if err != nil {
		logrus.Error(err)
		ctx.ViewData("msg", "读取数据出错")
		ctx.View("ui/sending.html")
		return
	}
	userId := ctx.GetCookie("userId")
	username := ctx.GetCookie("username")
	dto := services.RedEnvelopeSendingDTO{
		EnvelopeType: form.EnvelopeType,
		Username:     username,
		UserId:       userId,
		Blessing:     form.Blessing,
		Amount:       form.Amount,
		Quantity:     form.Quantity,
	}

	service := services.GetRedEnvelopeService()
	activity, err := service.SendOut(dto)
	if err != nil {
		logrus.Error(err)
		ctx.ViewData("msg", "发红包失败，系统出错")
		ctx.View("ui/sending.html")
		return
	}
	ctx.ViewData("activity", activity)
	ctx.Redirect("/envelope/home")
}
