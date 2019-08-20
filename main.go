package main

import (
	"fmt"
	"github.com/devfeel/dotweb"
)

type Card struct {
	Title   string
	Content string
	DocURL  string
	MpName  string
	Avatar  string
}

func main() {
	app := dotweb.New()
	app.SetDevelopmentMode()
	app.HttpServer.ServerFile("/static/*filepath", "./static/")
	app.HttpServer.Renderer().SetTemplatePath("./templates/")
	app.HttpServer.GET("/", func(context dotweb.Context) error {
		return context.WriteString("Hello World")
	})
	app.HttpServer.GET("/posts", IndexView)
	fmt.Println("Start on http://127.0.0.1:8080")
	app.StartServer(8080)
}

func IndexView(ctx dotweb.Context) error {
	url := "http://wx.qlogo.cn/mmhead/Q3auHgzwzM5Bp39F4nxFK6e8yMJ6Ju5SneIotBH1RptRJ8VBq8IZ4Q/0"
	c := Card{"标题", "内容", "链接", "微信公众号", url}
	lst := []Card{c, c}
	ctx.ViewData().Set("Cards", lst)
	err := ctx.View("index.html")
	return err
}

// todo: dbc
// todo: improve <a> style
// todo: posts API
