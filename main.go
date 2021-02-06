package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wcharczuk/go-chart/v2"
)

//Args to set some args
type Args struct {
	HistoryPath string
	OutputType  string
	Help        bool
	Number      int
}

func (args *Args) selfCheck() {
	outputType := args.OutputType
	if outputType != "all" && outputType != "json" && outputType != "png" {
		fmt.Println("args '-p' error. can use ' -h' to see how to use.")
		os.Exit(0)
	}
}

//InitArgs get the args
func InitArgs() *Args {
	args := new(Args)
	historyPathStr := "Set the History file `path`"
	outputTypeStr := "Set the `output` file type\n" +
		"There are three types to choose:\n" +
		"  json: output json\n" +
		"  png: output png\n" +
		"  all: output json and png"
	chartNumberStr := "Set the number of website in the chart"
	flag.StringVar(&args.HistoryPath, "p", "./History", historyPathStr)
	flag.StringVar(&args.OutputType, "o", "all", outputTypeStr)
	flag.IntVar(&args.Number, "n", 15, chartNumberStr)
	flag.BoolVar(&args.Help, "h", false, "show the `help`")
	flag.Parse()
	args.selfCheck()
	return args
}

type website struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func getHistoryData(historyPath string) []website {
	db, err := sql.Open("sqlite3", historyPath)
	if err != nil {
		fmt.Println("数据库连接失败")
		os.Exit(0)
	}

	rows, err := db.Query("select url from urls")
	if err != nil {
		fmt.Println("查询失败")
		os.Exit(0)
	}

	reg := regexp.MustCompile("https?://[^/]+")
	websiteMap := make(map[string]int)
	for rows.Next() {
		var url string
		rows.Scan(&url)
		webName := reg.FindString(url)
		if webName != "" {
			count, ok := websiteMap[webName]
			if ok {
				websiteMap[webName] = count + 1
			} else {
				websiteMap[webName] = 1
			}
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
	return websiteList
}

func outputJSON(websiteList []website) {
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
}

func outputChart(websiteList []website, chartNumber int) {
	if chartNumber > len(websiteList) {
		chartNumber = len(websiteList)
	}

	PieData := []chart.Value{}
	for i := 0; i < chartNumber; i++ {
		PieData = append(PieData, chart.Value{
			Label: strings.Split(websiteList[i].Name, "//")[1],
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

func outputWithArgs(websiteList []website, args *Args) {
	switch args.OutputType {
	case "all":
		outputJSON(websiteList)
		outputChart(websiteList, args.Number)
	case "json":
		outputJSON(websiteList)
	case "png":
		outputChart(websiteList, args.Number)
	}
}

func main() {
	args := InitArgs()
	fmt.Println(args.HistoryPath)
	if args.Help {
		flag.PrintDefaults()
		return
	}

	websiteList := getHistoryData(args.HistoryPath)
	outputWithArgs(websiteList, args)
}
