package views

import (
	"encoding/json"
	"fmt"
	"github.com/SAIKAII/skResk-Infra/httpclient"
	"github.com/SAIKAII/skResk-Infra/lb"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/SAIKAII/skResk-Envelope/services"

	"github.com/sirupsen/logrus"

	infra "github.com/SAIKAII/skResk-Infra"
	"github.com/SAIKAII/skResk-Infra/base"
	"github.com/SAIKAII/skResk-ui/core/users"
	"github.com/kataras/iris"
)

var (
	accountAppName  string
	envelopeAppName string
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

	accountAppName = fmt.Sprintf("http://%s/", base.Props().GetDefault("accountAppName", "skResk-Account"))
	envelopeAppName = fmt.Sprintf("http://%s/", base.Props().GetDefault("envelopeAppName", "skResk-Envelope"))
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
	body := strings.NewReader(fmt.Sprintf("userId=%s&offset=0&limit=200", userId))
	d := SendRequest("POST", "http://skResk-Envelope/listsent", body)
	if d == nil {
		logrus.Error("获取我发送的红包信息出错")
		return
	}

	var orders []services.RedEnvelopeGoodsDTO
	err := json.Unmarshal(d, &orders)
	if err != nil {
		logrus.Error(err)
		return
	}

	ctx.ViewData("orders", orders)
	ctx.ViewData("format", services.DefaultTimeFarmat)
	_ = ctx.View("ui/home.html")
}

// 我抢到的红包列表：recv_list.html /recvd/list
func (m *MobileView) receivedListHandler(ctx iris.Context) {
	userId := ctx.GetCookie("userId")
	body := strings.NewReader(fmt.Sprintf("userId=%s&offset=0&limit=200", userId))
	d := SendRequest("POST", "http://skResk-Envelope/listreceived", body)
	if d == nil {
		logrus.Error("获取我抢到的红包信息出错")
		return
	}

	var items []services.RedEnvelopeItemDTO
	err := json.Unmarshal(d, &items)
	if err != nil {
		logrus.Error(err)
		return
	}
	ctx.ViewData("items", items)
	ctx.ViewData("format", services.DefaultTimeFarmat)

	ctx.View("ui/recvd_list.html")
}

// 红包记录：re_one.html /list
func (m *MobileView) listHandler(ctx iris.Context) {
	envelopeNo := ctx.URLParamTrim("id")
	d := SendRequest("GET", fmt.Sprintf("http://skResk-Envelope/listorder?envelopeNo=%s", envelopeNo), nil)
	if d == nil {
		logrus.Error("获取单个商品信息出错")
		return
	}
	order := &services.RedEnvelopeGoodsDTO{}
	err := json.Unmarshal(d, order)
	if err != nil {
		logrus.Error(err)
		return
	}
	if order != nil {
		d = SendRequest("GET", fmt.Sprintf("http://skResk-Envelope/listitems?envelopeNo=%s", envelopeNo), nil)
		if d == nil {
			logrus.Error("获取抢红包用户信息出错")
			return
		}
		var items []services.RedEnvelopeItemDTO
		err = json.Unmarshal(d, &items)
		if err != nil {
			logrus.Error(err)
			return
		}
		totalAmount := decimal.NewFromFloat(0)
		for _, v := range items {
			totalAmount = totalAmount.Add(v.Amount)
		}
		t1, t2 := order.CreatedAt, order.UpdatedAt
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
	d := SendRequest("GET", fmt.Sprintf("http://skResk-Envelope/details?envelopeNo=%s", envelopeNo), nil)
	if d == nil {
		logrus.Error("获取红包详情出错")
		return
	}
	order := &services.RedEnvelopeGoodsDTO{}
	err := json.Unmarshal(d, order)
	if err != nil {
		logrus.Error(err)
		return
	}

	ctx.ViewData("order", order)
	ctx.ViewData("hasOrder", order != nil)
	ctx.ViewData("format", services.DefaultTimeFarmat)
	ctx.View("ui/re_details.html")
}

// 可抢红包：rev_home.html /rev/home
func (m *MobileView) receiveHomeHandler(ctx iris.Context) {
	body := strings.NewReader("offset=0&limit=200")
	d := SendRequest("POST", "http://skResk-Envelope/listreceviable", body)
	if d == nil {
		logrus.Error("获取可抢红包信息出错")
		return
	}
	var orders []services.RedEnvelopeGoodsDTO
	err := json.Unmarshal(d, &orders)
	if err != nil {
		logrus.Error(err)
		return
	}

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
	body := strings.NewReader(fmt.Sprintf("envelopeNo=%s&userId=%s&username=%s", envelopeNo, userId, username))
	d := SendRequest("POST", "http://skResk-Envelope/receive", body)
	if d == nil {
		logrus.Error("抢红包出错")
		return
	}
	item := &services.RedEnvelopeGoodsDTO{}
	err := json.Unmarshal(d, item)
	if err != nil {
		logrus.Error(err)
		return
	}

	msg := ""
	if d != nil {
		ctx.ViewData("hasReceived", true)
	} else {
		ctx.ViewData("hasReceived", false)
		msg = "抢红包失败"
	}

	d = SendRequest("GET", fmt.Sprintf("http://skResk-Envelope/listorder?envelopeNo=%s", envelopeNo), nil)
	if d == nil {
		logrus.Error(err)
		return
	}
	order := &services.RedEnvelopeGoodsDTO{}
	err = json.Unmarshal(d, order)
	if err != nil {
		logrus.Error(err)
		return
	}
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
	d, _ := json.Marshal(dto)
	body := strings.NewReader(string(d))
	d = SendRequest("POST", "http://skResk-Envelope/sendout", body)
	if d == nil {
		logrus.Error("发红包出错")
		return
	}
	activity := &services.RedEnvelopeActivity{}
	err = json.Unmarshal(d, activity)
	if err != nil {
		logrus.Error(err)
		ctx.ViewData("msg", "发红包失败，系统出错")
		ctx.View("ui/sending.html")
		return
	}
	ctx.ViewData("activity", activity)
	ctx.Redirect("/envelope/home")
}

func SendRequest(method, url string, body io.Reader) []byte {
	ec := base.EurekaClient()
	apps := &lb.Apps{Client: ec}
	c := httpclient.NewHttpClient(apps, nil)
	req, err := c.NewRequest(method, url, body, nil)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.Do(req)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	res.Body.Close()
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return d
}
