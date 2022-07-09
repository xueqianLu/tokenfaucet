package routers

import (
	"github.com/astaxie/beego"
	"github.com/hpb-project/tokenfaucet/common"
	"github.com/hpb-project/tokenfaucet/controllers"
)

func init() {
	control := &controllers.FaucetController{
		Networks: common.GetAllNetworkConfig(),
	}

	beego.Router("/", control, "get:Default")
	beego.Router("/api/tokenfaucet/v1/login/github", control, "get:GithubLoginHandler")
	beego.Router("/api/tokenfaucet/v1/login/github/callback", control, "get:GithubCallbackHandler")
	beego.Router("/api/tokenfaucet/v1/tokentransfer", control, "post:TokenTransfer")
}
