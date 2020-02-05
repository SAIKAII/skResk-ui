package public

import (
	"path/filepath"
	"runtime"
	"strings"

	infra "github.com/SAIKAII/skResk-Infra"
	"github.com/SAIKAII/skResk-Infra/base"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
)

type WebView struct {
}

func init() {
	infra.RegisterApi(&WebView{})
}

var mobileAgents = []string{"iphone", "android", "phone", "mobile", "wap", "netfront", "java", "opera mobi",
	"opera mini", "ucweb", "windows ce", "symbian", "series", "webos", "sony", "blackberry", "dopod",
	"nokia", "samsung", "palmsource", "xda", "pieplus", "meizu", "midp", "cldc", "motorola", "foma",
	"docomo", "up.browser", "up.link", "blazer", "helio", "hosin", "huawei", "novarra", "coolpad", "webos",
	"techfaith", "palmsource", "alcatel", "amoi", "ktouch", "nexian", "ericsson", "philips", "sagem",
	"wellcom", "bunjalloo", "maui", "smartphone", "iemobile", "spice", "bird", "zte-", "longcos",
	"pantech", "gionee", "portalmmm", "jig browser", "hiptop", "benq", "haier", "^lct", "320x320",
	"240x320", "176x220", "w3c ", "acs-", "alav", "alca", "amoi", "audi", "avan", "benq", "bird", "blac",
	"blaz", "brew", "cell", "cldc", "cmd-", "dang", "doco", "eric", "hipt", "inno", "ipaq", "java", "jigs",
	"kddi", "keji", "leno", "lg-c", "lg-d", "lg-g", "lge-", "maui", "maxo", "midp", "mits", "mmef", "mobi",
	"mot-", "moto", "mwbp", "nec-", "newt", "noki", "oper", "palm", "pana", "pant", "phil", "play", "port",
	"prox", "qwap", "sage", "sams", "sany", "sch-", "sec-", "send", "seri", "sgh-", "shar", "sie-", "siem",
	"smal", "smar", "sony", "sph-", "symb", "t-mo", "teli", "tim-", "tosh", "tsm-", "upg1", "upsi", "vk-v",
	"voda", "wap-", "wapa", "wapi", "wapp", "wapr", "webc", "winw", "winw", "xda", "xda-",
	"Googlebot-Mobile"}

func (w *WebView) Init() {
	app := base.Iris()
	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(f)
	views := iris.HTML(dir, ".html")
	views.Reload(true)
	views.Layout("ui/layouts/layout.html")
	app.RegisterView(views)
	app.Any("/", indexHandler)
	app.Any("", indexHandler)
}

func indexHandler(ctx iris.Context) {
	ua := ctx.GetHeader("user-agent")
	ua = strings.ToLower(ua)
	isMobile := false
	logrus.Info(ua)
	for _, mobileAgent := range mobileAgents {
		if strings.Index(ua, mobileAgent) >= 0 {
			isMobile = true
			break
		}
	}

	if isMobile {
		ctx.Redirect("/home")
	} else {
		ctx.Redirect("/home")
	}
}
