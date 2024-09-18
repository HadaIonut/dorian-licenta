package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SensorData struct {
	Voltage   float64 `json:"voltage"`
	Current   float64 `json:"current"`
	PanelName string  `json:"panelName"`
}

type DataToPlot struct {
	Voltage   float64 `json:"voltage"`
	Current   float64 `json:"current"`
	Timestamp float64 `json:"timestamp"`
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.POST("/uploadData", uploadData)
	r.GET("/data/:panelName", getData)
	r.GET("/", getDashboard)
	r.Run()
}

func uploadData(c *gin.Context) {
	var body SensorData
	var content string

	err := c.BindJSON(&body)

	if err != nil {
		fmt.Println(err)
		return
	}

	filePath := "data/" + body.PanelName + ".csv"
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		content = "voltage,current,timestamp"
	} else {
		data, err := ioutil.ReadFile(filePath)
		content += string(data)

		if err != nil {
			fmt.Println(err)
			return
		}
	}

	now := time.Now()
	content += "\n " + strconv.FormatFloat(body.Voltage, 'f', 2, 64) + "," + strconv.FormatFloat(body.Current, 'f', 2, 64) + "," + now.Format("2006-01-02T15:04:05") + "+03:00"
	ioutil.WriteFile(filePath, []byte(content), 0777)

	c.Status(http.StatusOK)
}

type DataResponse struct {
	Data []map[string]interface{} `json:"data"`
}

func getData(c *gin.Context) {
	var response DataResponse
	fileName := c.Param("panelName")

	parsedData, err := Csv2Json("data/" + fileName + ".csv")
	response.Data = parsedData

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, response)
}

func getDashboard(c *gin.Context) {
	var files []string

	entries, err := os.ReadDir("./data")
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		files = append(files, strings.Split(e.Name(), ".")[0])
	}
	c.HTML(http.StatusOK, "index.html", gin.H{"files": files})
}
