package main

import (
	infra "github.com/SAIKAII/skResk-Infra"
	"github.com/tietang/props/consul"
	"github.com/tietang/props/ini"
	"github.com/tietang/props/kvs"
)

func main() {
	file := kvs.GetCurrentFilePath("boot.ini", 1)
	conf := ini.NewIniFileCompositeConfigSource(file)
	conf.Set("profile", "dev")
	conf.Set("cfgName", "skResk")
	addr := conf.GetDefault("consul.address", "127.0.0.1:8500")
	contexts := conf.KeyValue("consul.contexts").Strings()
	consulConf := consul.NewCompositeConsulConfigSourceByType(contexts, addr, kvs.ContentIni)
	consulConf.Add(conf)
	app := infra.New(consulConf)
	app.Start()
}
