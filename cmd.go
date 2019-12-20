package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"golang.org/x/exp/errors/fmt"
	"golang.org/x/exp/rand"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//Remove this and make it argument
const token = "XXX" //Your Scalyr token here
const scalyrUrl string = "https://app.scalyr.com/api/query"
var skipKeyWords = []string{"password", "otp", "reminder", "verification", "redacted", "zomato", "imo", "order for", "verification Code"}


var starttime int64 = 0
var endtime int64 = 0
var countOfSkippedText int64 = 0
var countOfRepeatedText int64 = 0
var totalScanned int64 = 0

var messageTexts map[string]string

type scalyrResponse struct{
	Matches []match `json:"matches"`
	ContinuationToken string `json:"continuationToken"`
}

type match struct{
	Attributes attrib `json:"attributes"`
	Severity string `json:"severity"`
}

type attrib struct{
	Message2 string `json:"message_"`
}

/**
Check if hour to deduct from current time provided as argument, else do random.
 */
func main(){
	fmt.Println("Starting...")
	starttimeforpgm :=  time.Now().Unix()
	var hour int = -1
	if(len(os.Args) > 1){
		hour,err := strconv.Atoi(os.Args[1])
		if(err != nil || hour < 1 || hour > 72){
			fmt.Println("Error, argument should be between 1 - 72")
		}
	}
	rand.Seed(uint64(time.Now().UnixNano()))
	messageTexts = make(map[string]string)
	if(hour == -1){
		hour = rand.Intn(24)
	}
	endtime = ((time.Now().Unix()  - (int64) ((hour) * 60 * 60)))*1000
	starttime = endtime - (6 * 60 * 60 * 1000)
	fmt.Println("Getting text between time (epoch seconds)", starttime, endtime)
	getContentForPeriod("", 0)
	texts := [][]string{{"#", "Text", "Category"}}
	var i int64 = 0
	for _, text := range messageTexts{
		i++
		texts = append(texts, []string{"" + strconv.FormatInt(i, 10), text, ""})
	}
	fileName := writeCSVFile(texts)
	fmt.Println("CSV File written", fileName)
	fmt.Printf("TotalScanned = %d , Repeated = %d , Skipped(keyword) = %d , TotalFound = %d \n", totalScanned, countOfRepeatedText, countOfSkippedText, len(messageTexts))
	endtimeforpgm := time.Now().Unix()
	fmt.Println("Time taken in secs", (endtimeforpgm - starttimeforpgm))
}

/**
Call Scalyr api to get log data for the filter, paginated recursive query
 */
func getContentForPeriod(continuationToken string, i int){
	url := scalyrUrl
	fmt.Printf("Taking %d set of 1000 message text\n", (i+1))

	var jsonStr = []byte(`{"token":"` +  token + `",`  + `"queryType":"log",`  + `"filter":"text SMS queued to SharQ with params", 
	"startTime": ` + strconv.FormatInt(starttime, 10) + `,
  	"endTime":  ` + strconv.FormatInt(endtime, 10) + `,
  	"maxCount": 1000,
	"continuationToken":  "` + continuationToken + `",
  	"pageMode":  "head"` + `}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var target scalyrResponse
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	json.Unmarshal([]byte(bodyString), &target)
	fmt.Println("Next page token", target.ContinuationToken)
	for _, match := range target.Matches{
		text := strings.Replace(match.Attributes.Message2, "SMS queued to SharQ with params -", "", 1)
		msgText := getMessageText(text)
		lowerText := strings.ToLower(msgText)
		flag := false;
		totalScanned++
		for _, str := range skipKeyWords{
			if strings.Contains(lowerText, str){
				//contains skip words
				countOfSkippedText++
				flag = true
				break
			}
		}
		if(!flag){
			if _, ok := messageTexts[msgText]; ok {
				//nothing to do here, repeated text
				countOfRepeatedText++
			} else{
				messageTexts[msgText] = msgText
			}
		}
	}
	fmt.Println("Total texts", len(messageTexts))
	if(len(messageTexts) < 10000 || target.ContinuationToken == ""){
		getContentForPeriod(target.ContinuationToken, (i+1))
	}
}


/**
Get Message text from JSON
*/
func getMessageText(text string) string{
	actualText := ""
	//fmt.Println("text", text)
	i := 0;
	for _, char := range text {
		if(char == 'u' && i== 0 ) {
			i++
		} else if(char == '\'' && i== 1){
			i++
		} else if(char == '\'' && i== 2){
			i++
		} else if(char == ',' && i== 3){
			i++
		} else if(i == 2){
			actualText += string(char)
		} else if(i == 4 ){
			break
		} else{
			//skip
		}

	}
	return actualText
}

/**
Write as CSV file
 */
func writeCSVFile(rows [][]string) string{
	fmt.Println("In writeCSVFile")
	t1 := time.Now().Unix()
	fileName := "messagetext_" + strconv.FormatInt(t1, 10) + ".csv"
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range rows {
		err := writer.Write(row)
		if err != nil {
			log.Fatal("Cannot write now", err)
		}
	}
	return fileName
}




