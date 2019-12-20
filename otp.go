package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"golang.org/x/exp/errors/fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//Remove this and make it argument
const authToken = "XXX" //Your scalyr read token
const url string = "https://app.scalyr.com/api/query"
var interestWords = []string{"otp", "verification code"}


var start int64 = 0
var end int64 = 0
var total int64 = 0
var matched int64 = 0

var authIds map[string]string

type response struct{
	Matches []matcher `json:"matches"`
	ContinuationToken string `json:"continuationToken"`
}

type matcher struct{
	Attributes attribute `json:"attributes"`
	Severity string `json:"severity"`
}

type attribute struct{
	AuthId string `json:"auth_id"`
	Message2 string `json:"message_"`
}

/**
Check if hour to deduct from current time provided as argument, else do random.
 */
func main(){
	fmt.Println("Starting...")
	starttimeforpgm :=  time.Now().Unix()
	var hour int = 1
	if(len(os.Args) > 1){
		hour,err := strconv.Atoi(os.Args[1])
		if(err != nil || hour < 1 || hour > 24){
			fmt.Println("Error, argument should be between 1 - 72")
		}
	}
	start = (time.Now().Unix()  - (int64)  (hour * 60 * 60))*1000
	end = (time.Now().Unix()  - (int64) ((hour - 1) * 60 * 60)) * 1000
	fmt.Println("Getting text between time (epoch seconds)", start, end)
	authIds = make(map[string]string)
	getAllContentForPeriod("", 0)
	authTokens := [][]string{{"#", "AuthId", "Sample"}}
	var i int64 = 0
	for k, v := range authIds{
		i++
		authTokens = append(authTokens, []string{"" + strconv.FormatInt(i, 10), k, v})
	}
	fileName := writeCSVFileForAuths(authTokens)
	fmt.Println("CSV File written", fileName)
	fmt.Printf("TotalScanned = %d , Matched = %d, TotalFound = %d \n", total, matched, len(authTokens) - 1)
	endtimeforpgm := time.Now().Unix()
	fmt.Println("Time taken in secs", (endtimeforpgm - starttimeforpgm))
}

/**
Call Scalyr api to get log data for the filter, paginated recursive query
 */
func getAllContentForPeriod(continuationToken string, i int){
	url := url
	fmt.Printf("Taking %d set of 1000 message text\n", (i+1))

	var jsonStr = []byte(`{"token":"` +  authToken + `",`  + `"queryType":"log",`  + `"filter":"text SMS queued to SharQ with params", 
	"startTime": ` + strconv.FormatInt(start, 10) + `,
  	"endTime":  ` + strconv.FormatInt(end, 10) + `,
  	"maxCount": 5000,
	"continuationToken":  "` + continuationToken + `",
  	"pageMode":  "head"` + `}`)
	//fmt.Println("url", string(jsonStr))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var target response
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	json.Unmarshal([]byte(bodyString), &target)
	//fmt.Println("Next page token", target.ContinuationToken)
	for _, match := range target.Matches{
		text := strings.Replace(match.Attributes.Message2, "SMS queued to SharQ with params -", "", 1)
		msgText, authId := getMessageTextAndAuthToken(text, match)
		lowerText := strings.ToLower(msgText)
		flag := false;
		for _, str := range interestWords{
			if strings.Contains(lowerText, str){
				flag = true
				matched++
				break
			}
		}
		if(flag){
			if _, ok := authIds[authId]; ok {
				//skip, already present
			} else{
				authIds[authId] = msgText
			}
		} else{

		}
		total++
	}
	//fmt.Println("Total scanned", total, matched)
	if(target.ContinuationToken != "" && i < 100){
		getAllContentForPeriod(target.ContinuationToken, (i+1))
	}
}


/**
Get Message text from JSON
*/
func getMessageTextAndAuthToken(text string, match matcher) (string, string){
	actualText := ""
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
	return actualText, match.Attributes.AuthId
}

/**
Write as CSV file
 */
func writeCSVFileForAuths(rows [][]string) string{
	fmt.Println("In writeCSVFileForAuths")
	t1 := time.Now().Unix()
	fileName := "Auths_" + strconv.FormatInt(t1, 10) + ".csv"
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




