package routers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/hpb-project/tokenfaucet/common"
	"github.com/hpb-project/tokenfaucet/controllers"
)

func init() {
	control := &controllers.FaucetController{
		Networks: common.GetAllNetworkConfig(),
	}
	fmt.Println("route start")

	beego.Router("/tokenfaucet/api/v1/tokentransfer", control, "post:TokenTransfer")
}
