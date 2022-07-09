package main

import (
	"github.com/astaxie/beego"
	_ "github.com/hpb-project/tokenfaucet/cacheTools"
	_ "github.com/hpb-project/tokenfaucet/routers"
)

func main() {
	beego.Run()
}
