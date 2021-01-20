package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wcharczuk/go-chart/v2"
)

type website struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func main() {
	db, err := sql.Open("sqlite3", "./History")
	if err != nil {
		fmt.Println("数据库连接失败")
		return
	}

	rows, err := db.Query("select url from urls")
	if err != nil {
		fmt.Println("查询失败")
		return
	}

	reg := regexp.MustCompile("https?://[^/]*")
	websiteMap := make(map[string]int)
	for rows.Next() {
		var url string
		rows.Scan(&url)
		website := reg.FindString(url)
		count, ok := websiteMap[website]
		if ok {
			websiteMap[website] = count + 1
		} else {
			websiteMap[website] = 1
		}
	}

	var websiteList []website
	for key, value := range websiteMap {
		websiteList = append(websiteList, website{
			Name:  key,
			Count: value,
		})
	}
	sort.Slice(websiteList, func(i, j int) bool { return websiteList[i].Count > websiteList[j].Count })

	historyJSON, err := json.Marshal(websiteList)
	if err != nil {
		fmt.Println("json 转换错误")
	} else {
		jsonFile, err := os.Create("history.json")
		if err != nil {
			fmt.Println("创建json错误")
		} else {
			jsonFile.Write(historyJSON)
			jsonFile.Close()
		}
	}

	PieData := []chart.Value{}
	for i := 0; i < 15; i++ {
		PieData = append(PieData, chart.Value{
			Label: websiteList[i].Name,
			Value: float64(websiteList[i].Count),
		})
	}

	pie := chart.DonutChart{
		Width:  2000,
		Height: 2000,
		DPI:    100,
		Values: PieData,
	}

	pieChart, err := os.Create("history.png")
	if err != nil {
		fmt.Println("输出图片错误")
	} else {
		pie.Render(chart.PNG, pieChart)
		pieChart.Close()
	}

}
