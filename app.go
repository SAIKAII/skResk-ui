package main

import (
	_ "github.com/SAIKAII/skResk-Account/core/accounts"
	_ "github.com/SAIKAII/skResk-Envelope/apis/web"
	_ "github.com/SAIKAII/skResk-Envelope/core/envelopes"
	"github.com/SAIKAII/skResk-Infra"
	"github.com/SAIKAII/skResk-Infra/base"
	_ "github.com/SAIKAII/skResk-ui/public"
	_ "github.com/SAIKAII/skResk-ui/public/views"
	_ "github.com/SAIKAII/skResk-ui/views"
)

func init() {
	infra.Register(&base.PropsStarter{})
	infra.Register(&base.DbxDatabaseStarter{})
	infra.Register(&base.ValidatorStarter{})
	infra.Register(&infra.WebApiStarter{})
	infra.Register(&base.HookStarter{})
	infra.Register(&base.EurekaStarter{})
	infra.Register(&base.IrisServerStarter{})
}
