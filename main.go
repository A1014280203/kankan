package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/devfeel/dotweb"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"time"
)

type Card struct {
	Title      string
	Content    string
	DocURL     string
	MpName     string
	Avatar     string
	CreateTime uint32
	// not from db
	TimeDiff string
}

var db *sql.DB

func main() {
	app := dotweb.New()
	//set debug mode
	//app.SetDevelopmentMode()
	//option := cors.NewConfig().UseDefault()

	db, _ = sql.Open("mysql", "user:password@tcp(ip:port)/weread?charset=utf8")
	defer db.Close()

	app.HttpServer.ServerFile("/static/*filepath", "./static/")
	app.HttpServer.Renderer().SetTemplatePath("./templates/")
	app.HttpServer.GET("/", IndexView)
	app.HttpServer.GET("/recom", recomView)
	app.HttpServer.POST("/recom", recomRecv)
	app.HttpServer.GET("/posts", MorePost)
	app.HttpServer.GET("/mp", searchMp)
	app.HttpServer.GET("/h", func(context dotweb.Context) error {
		return context.WriteString("Hello World")
	})
	fmt.Println("Start on http://127.0.0.1:8080")
	err := app.StartServer(8080)
	if err != nil {
		panic(err.Error())
	}
}

func IndexView(ctx dotweb.Context) error {
	/* set cards and last timestamp to html */
	nowTimeStamp := uint32(time.Now().Unix())
	rows := dbcQueryPost(nowTimeStamp)
	defer rows.Close()
	var cards []Card
	var followTime uint32
	for rows.Next() {
		var card Card
		err := rows.Scan(&card.Title, &card.Content, &card.DocURL, &card.MpName, &card.Avatar, &card.CreateTime)
		if err != nil {
			panic(err.Error())
		}
		card.TimeDiff = timeDiff(int64(nowTimeStamp), int64(card.CreateTime))
		followTime = card.CreateTime
		cards = append(cards, card)
	}
	ctx.ViewData().Set("cards", cards)
	ctx.ViewData().Set("followTime", followTime)
	ctx.AddView("layout.html")
	err := ctx.View("index.html")
	return err
}

func MorePost(ctx dotweb.Context) error {
	/* return cards and last timestamp as JSON to user */
	nowTimeStamp := uint32(time.Now().Unix())
	timeString := ctx.QueryString("followTime")
	followTime64, err := strconv.ParseUint(timeString, 10, 32)
	followTime := uint32(followTime64)
	if err != nil {
		panic(err.Error())
	}

	rows := dbcQueryPost(followTime)
	defer rows.Close()
	var cards []Card
	for rows.Next() {
		var card Card
		err := rows.Scan(&card.Title, &card.Content, &card.DocURL, &card.MpName, &card.Avatar, &card.CreateTime)
		if err != nil {
			panic(err.Error())
		}
		card.TimeDiff = timeDiff(int64(nowTimeStamp), int64(card.CreateTime))
		followTime = card.CreateTime
		cards = append(cards, card)
	}

	data := make(map[string]interface{})
	data["cards"] = cards
	data["followTime"] = followTime
	b, err := json.Marshal(data)
	if err != nil {
		panic(err.Error())
	}
	err = ctx.WriteJsonBlob(b)

	return err
}

func recomView(ctx dotweb.Context) error {
	ctx.AddView("layout.html")
	return ctx.View("recom.html")
}

func recomRecv(ctx dotweb.Context) error {
	MpName := ctx.PostFormValue("MpName")
	ShareURL := ctx.PostFormValue("ShareURL")
	Reason := ctx.PostFormValue("Reason")
	Ip := ctx.RemoteIP()
	nowTimeStamp := uint32(time.Now().Unix())
	dbcInsertrecom(MpName, ShareURL, Reason, Ip, nowTimeStamp)
	return recomView(ctx)
}

func searchMp(ctx dotweb.Context) error {
	MpName := ctx.QueryString("name")
	nowTimeStamp := uint32(time.Now().Unix())

	rows := dbcQueryPostByMp(nowTimeStamp, MpName)
	defer rows.Close()
	var cards []Card
	for rows.Next() {
		var card Card
		err := rows.Scan(&card.Title, &card.Content, &card.DocURL, &card.MpName, &card.Avatar, &card.CreateTime)
		if err != nil {
			panic(err.Error())
		}
		card.TimeDiff = timeDiff(int64(nowTimeStamp), int64(card.CreateTime))
		cards = append(cards, card)
	}
	if len(cards) > 0 {
		ctx.ViewData().Set("cards", cards)
		ctx.ViewData().Set("MpName", MpName)
		ctx.ViewData().Set("postCount", len(cards))
	}
	ctx.AddView("layout.html")
	err := ctx.View("result.html")
	return err
}

func dbcQueryPost(followTime uint32) *sql.Rows {
	aliasCol := "title as Title, content as Content, doc_url as DocURL, mp_name as MpName, avatar as Avatar, createTime as CreateTime"
	stat, err := db.Prepare(fmt.Sprintf("SELECT %s FROM post WHERE createTime < ? ORDER BY createTime DESC limit 10", aliasCol))
	defer stat.Close()
	if err != nil {
		panic(err.Error())
	}
	rows, err := stat.Query(followTime)
	if err != nil {
		panic(err.Error())
	}

	return rows
}

func dbcInsertrecom(MpName, ShareURL, Reason, Ip string, nowTimeStamp uint32) {
	stat, err := db.Prepare("INSERT INTO recom VALUES (?, ?, ?, ?, ?)")
	defer stat.Close()
	if err != nil {
		panic(err.Error())
	}
	_, err = stat.Exec(MpName, ShareURL, Reason, Ip, nowTimeStamp)
	if err != nil {
		panic(err.Error())
	}
}

func dbcQueryPostByMp(followTime uint32, MpName string) *sql.Rows {
	aliasCol := "title as Title, content as Content, doc_url as DocURL, mp_name as MpName, avatar as Avatar, createTime as CreateTime"
	stat, err := db.Prepare(fmt.Sprintf("SELECT %s FROM post WHERE createTime < ? and mp_name = ? ORDER BY createTime DESC limit 25", aliasCol))
	defer stat.Close()
	if err != nil {
		panic(err.Error())
	}
	rows, err := stat.Query(followTime, MpName)
	if err != nil {
		panic(err.Error())
	}

	return rows
}

func timeDiff(t1, t2 int64) string {
	d1 := time.Unix(t1, 0)
	d2 := time.Unix(t2, 0)
	diff := d1.Sub(d2)
	diffString := timeDiffToString(diff)
	return diffString
}

func timeDiffToString(d time.Duration) string {
	switch {
	case d.Hours() >= 24*30:
		return fmt.Sprintf("%d月之前", int(d.Hours()/(24*30)))
	case d.Hours() >= 24:
		return fmt.Sprintf("%d天之前", int(d.Hours()/24))
	case d.Hours() >= 1:
		return fmt.Sprintf("%d小时之前", int(d.Hours()))
	case d.Minutes() >= 1:
		return fmt.Sprintf("%d分钟之前", int(d.Minutes()))
	default:
		return fmt.Sprintf("%d秒之前", int(d.Seconds()))
	}
}

/*

index init the end timestamp

restAPI follow the timestamp index set

re-request index will re-init the end timestamp

*/
