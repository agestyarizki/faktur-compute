package controller

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type Invoice struct {
	CustomerName string     `json:"customer_name"`
	NamaBarang   string     `json:"nama_barang"`
	Date         *time.Time `json:"date"`
	Total        float64    `json:"total"`
}

type FinalSummaryInvoice struct {
	TotalInvoice float64   `json:"total_invoice"`
	ListInvoice  []Invoice `json:"list_invoice"`
}

type FinalInvoicePermonth struct {
	TotalInvoicePermonth   float64               `json:"total_invoice_permonth"`
	DataListSummaryInvoice []FinalSummaryInvoice `json:"data_list_summary_invoice"`
}

func ComputeData(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Println(err.Error())
		ErrorReturnsWithJson(c, err.Error())
		return
	}

	basePathTemp := "temp"
	if _, err := os.Stat(basePathTemp); os.IsNotExist(err) {
		os.MkdirAll(basePathTemp, 0755)
	}

	unixTime := time.Now().UnixNano() / 1000000
	filenameExc := fmt.Sprintf("computeData_%d.xlsx", unixTime)
	path := filepath.Join(basePathTemp, filenameExc)
	if err := c.SaveUploadedFile(file, path); err != nil {
		log.Println(err.Error())
		ErrorReturnsWithJson(c, err.Error())
		return
	}

	xlsx, err := excelize.OpenFile(path)
	if err != nil {
		log.Println(err.Error())
		ErrorReturnsWithJson(c, err.Error())
		return
	}

	rows := xlsx.GetRows("Sheet1")

	var dataComputasi = make(map[string][]Invoice)
	var dataFinal = make(map[string][]FinalSummaryInvoice)

	for index, dataRow := range rows {
		if index > 0 {
			timeParse := ParseFloatData(dataRow[0])
			dateTime := timeFromExcelTime(timeParse, false)
			timeFormat := dateTime.Format(time.RFC3339)
			dataKey := timeFormat + " - " + ReplaceWhiteSpace(dataRow[1])
			dataComputasi[dataKey] = append(dataComputasi[dataKey], Invoice{
				CustomerName: dataRow[1],
				NamaBarang:   dataRow[2],
				Total:        ParseFloatData(dataRow[3]),
				Date:         &dateTime,
			})
		}
	}

	for _, eachInvoice := range dataComputasi {
		var havePackage = false
		var totalInvoice float64
		var monthData string
		for indexKey, eachData := range eachInvoice {
			isPackage, err := regexp.MatchString("Paket", eachData.NamaBarang)
			if err != nil {
				log.Println(err.Error())
				ErrorReturnsWithJson(c, err.Error())
				return
			}

			if indexKey == 0 {
				monthData = eachData.Date.Format("2006-01")
			}

			if isPackage == true {
				if havePackage == false {
					havePackage = true
					totalInvoice = eachData.Total
				}
			} else {
				if len(eachInvoice) > 1 {
					totalInvoice = totalInvoice + eachData.Total
				}
			}
		}

		//totalAllInvoice = totalAllInvoice + totalInvoice
		dataFinal[monthData] = append(dataFinal[monthData], FinalSummaryInvoice{
			TotalInvoice: totalInvoice,
			ListInvoice:  eachInvoice,
		})
	}

	// EACH DATA PER MONTH
	var finalDataPermonth = make(map[string]FinalInvoicePermonth)
	for index, data := range dataFinal {
		var totalInvoicePermonth float64
		for _, eachInvoice := range data {
			totalInvoicePermonth = eachInvoice.TotalInvoice + totalInvoicePermonth
		}

		finalDataPermonth[index] = FinalInvoicePermonth{
			TotalInvoicePermonth:   totalInvoicePermonth,
			DataListSummaryInvoice: data,
		}
	}

	c.JSON(http.StatusOK, finalDataPermonth)
	return
}

func ErrorReturnsWithJson(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    http.StatusInternalServerError,
		"message": message,
	})
}

func ParseFloatData(num string) float64 {
	numFloat, err := strconv.ParseFloat(num, 64)
	if err != nil {
		log.Println(err)
	}

	return numFloat
}

func ReplaceWhiteSpace(words string) string {
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(words, " ")
}
