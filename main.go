package main

import (
	"bufio"
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

fjsdkfsdjkdf
type SensorOutput struct {
	Voltage float64 `json:"voltage"`
	Current float64 `json:"current"`
}

type UploadMessage struct {
	PanelName string       `json:"panelName"`
	Data1     SensorOutput `json:"data1"`
	Data2     SensorOutput `json:"data2"`
	Data3     SensorOutput `json:"data3"`
	Data4     SensorOutput `json:"data4"`
	Data5     SensorOutput `json:"data5"`
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
	var body UploadMessage
	var content string

	err := c.BindJSON(&body)

	if err != nil {
		fmt.Println(err)
		return
	}

	filePath := "data/" + body.PanelName + ".csv"
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		content = "data1_voltage,data1_current,data2_voltage,data2_current,data3_voltage,data3_current,data4_voltage,data4_current,data5_voltage,data5_current,timestamp"
	} else {
		data, err := ioutil.ReadFile(filePath)
		content += string(data)

		if err != nil {
			fmt.Println(err)
			return
		}
	}

	now := time.Now()
	content += "\n" +
		strconv.FormatFloat(body.Data1.Voltage, 'f', 2, 64) + "," + strconv.FormatFloat(body.Data1.Current, 'f', 2, 64) + "," +
		strconv.FormatFloat(body.Data2.Voltage, 'f', 2, 64) + "," + strconv.FormatFloat(body.Data2.Current, 'f', 2, 64) + "," +
		strconv.FormatFloat(body.Data3.Voltage, 'f', 2, 64) + "," + strconv.FormatFloat(body.Data3.Current, 'f', 2, 64) + "," +
		strconv.FormatFloat(body.Data4.Voltage, 'f', 2, 64) + "," + strconv.FormatFloat(body.Data4.Current, 'f', 2, 64) + "," +
		strconv.FormatFloat(body.Data4.Voltage, 'f', 2, 64) + "," + strconv.FormatFloat(body.Data5.Current, 'f', 2, 64) + "," +
		now.Format("2006-01-02T15:04:05") + "+03:00"

	ioutil.WriteFile(filePath, []byte(content), 0777)

	c.Status(http.StatusOK)
}

type DataResponse struct {
	Data []Entry `json:"data"`
}

type Entry struct {
	Timestamp string       `json:"timestamp"`
	Data1     SensorOutput `json:"data1"`
	Data2     SensorOutput `json:"data2"`
	Data3     SensorOutput `json:"data3"`
	Data4     SensorOutput `json:"data4"`
	Data5     SensorOutput `json:"data5"`
}

func getData(c *gin.Context) {
	var response DataResponse
	fileName := c.Param("panelName")

	parsedData, err := Csv2Json("data/" + fileName + ".csv")

	for _, line := range parsedData {
		var entry Entry
		entry.Timestamp = line["timestamp"].(string)

		entry.Data1.Current, _ = strconv.ParseFloat(line["data1_current"].(string), 64)
		entry.Data1.Voltage, _ = strconv.ParseFloat(line["data1_voltage"].(string), 64)
		entry.Data2.Current, _ = strconv.ParseFloat(line["data2_current"].(string), 64)
		entry.Data2.Voltage, _ = strconv.ParseFloat(line["data2_voltage"].(string), 64)
		entry.Data3.Current, _ = strconv.ParseFloat(line["data3_current"].(string), 64)
		entry.Data3.Voltage, _ = strconv.ParseFloat(line["data3_voltage"].(string), 64)
		entry.Data4.Current, _ = strconv.ParseFloat(line["data4_current"].(string), 64)
		entry.Data4.Voltage, _ = strconv.ParseFloat(line["data4_voltage"].(string), 64)
		entry.Data5.Current, _ = strconv.ParseFloat(line["data5_current"].(string), 64)
		entry.Data5.Voltage, _ = strconv.ParseFloat(line["data5_voltage"].(string), 64)

		response.Data = append(response.Data, entry)
	}

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, response)
}

func readLastLine(filename string) (string, error) {
	readFile, err := os.Open(filename)

	if err != nil {
		return "", err
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}
	return fileLines[len(fileLines)-1], nil
}

type FileWithInfo struct {
	Content string
	HasErr  bool
}

func datasetHasErrors(path string) bool {
	line, err := readLastLine(path)
	line = strings.Trim(line, " ")

	if err != nil {
		panic(err)
	}

	values := strings.Split(line, ",")

	for _, value := range values {
		intVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
		if intVal == 0.0 {
			return true
		}
	}

	return false
}

func getDashboard(c *gin.Context) {
	var files []FileWithInfo

	entries, err := os.ReadDir("./data")
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		var curFile FileWithInfo
		curFile.Content = strings.Split(e.Name(), ".")[0]
		curFile.HasErr = datasetHasErrors("./data/" + e.Name())

		files = append(files, curFile)
	}
	c.HTML(http.StatusOK, "index.html", gin.H{"files": files})
}
