package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func checkErrors(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

//func generatePostRequestData() url.Values {
//	loginRequest := url.Values{}
//	loginRequest.Set("__EVENTTARGET", "")
//	loginRequest.Set("__EVENTARGUMENT", "")
//	loginRequest.Set("__VIEWSTATE", "/wEPDwULLTE1NjMxNjMxNzFkZCo3T3kAnddTDFryr26qaofiTp5p")
//	loginRequest.Set("__VIEWSTATEGENERATOR", "C2EE9ABB")
//	loginRequest.Set("__EVENTVALIDATION", "/wEdAAYNcEy/uvEwBm4by+oKLWkjSvD5Cbpu3w0ab2H9f5rbFEPTPkdPWl+8YN2NtDCtxifN+DvxnwFeFeJ9MIBWR693w+qCzNvQHKCQwl8+YzOKE62xJNKuHibH70Ul6qoa4F8sDaR1uxEyo1xbP9xcXI4vvNcYtQ==")
//	loginRequest.Set("TxtEmail", "id")
//	loginRequest.Set("txtPass", "pass")
//	loginRequest.Set("Button1", "Login")
//	loginRequest.Set("txtEmailId", "")
//	return loginRequest
//}

func setHeaders(req *http.Request, length int64, cookies string, referrer string) {
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,gu;q=0.8,hi;q=0.7")
	//req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Length", strconv.FormatInt(length, 10))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Cookie", cookies)
	req.Header.Set("DNT", "1")
	req.Header.Set("Host", "psd.bits-pilani.ac.in")
	req.Header.Set("Origin", "http://psd.bits-pilani.ac.in")
	req.Header.Set("Referer", referrer)
	//req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Mobile Safari/537.36)")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

}

func decodeJSON(response *io.ReadCloser) []map[string]interface{} {
	decoder := json.NewDecoder(*response)
	var array map[string]string
	err := decoder.Decode(&array)
	checkErrors(err)
	decoder = json.NewDecoder(strings.NewReader(array["d"]))
	var dataArray []map[string]interface{}
	err = decoder.Decode(&dataArray)
	checkErrors(err)

	return dataArray
}

func findInMapArray(mapArray []map[string]interface{}, key string, value interface{}) map[string]interface{} {
	for _, object := range mapArray {
		if val, ok := object[key]; ok && val == value {
			return object
		}
	}
	return nil
}

func mapToString(mapToConvert map[string]interface{}) string {
	var result string
	for key, value := range mapToConvert {
		result += fmt.Sprintln(key + ": " + fmt.Sprintf("%v", value))
	}
	return result
}

func writeCSV(stationList, problemBank []map[string]interface{}) {
	//Open file for csv
	csvFile, err := os.Create("StationDetails.csv")
	checkErrors(err)
	csvWriter := csv.NewWriter(csvFile)

	csvData := make([]string, 10)
	//Head
	csvData[0] = "Station ID"
	csvData[1] = "Company Name"
	csvData[2] = "Location"
	csvData[3] = "Industry Domain"
	csvData[4] = "Preferred Branches"
	csvData[5] = "Stipend (UG)"
	csvData[6] = "Stipend (PG)"
	csvData[7] = "Facilities (Raw)"
	csvData[8] = "Projects"
	csvData[9] = "Have Accommodation?"
	err = csvWriter.Write(csvData)
	checkErrors(err)
	var length = (int64)(len(stationList))
	for i, station := range stationList {
		problemBankCounterpart := findInMapArray(problemBank, "StationId", station["StationId"])
		projectAndFacilitiesCounterpart := getStationDetails(fmt.Sprintf("%v", problemBankCounterpart["StationId"]), fmt.Sprintf("%v", problemBankCounterpart["CompanyId"]))
		if problemBankCounterpart != nil {
			csvData[0] = fmt.Sprintf("%v", station["StationId"])
			csvData[1] = fmt.Sprintf("%v", problemBankCounterpart["CompanyName"])
			csvData[2] = fmt.Sprintf("%v", problemBankCounterpart["City"])
			csvData[3] = fmt.Sprintf("%v", problemBankCounterpart["IndustryDomain"])
			csvData[4] = fmt.Sprintf("%v", problemBankCounterpart["Tags"])
			csvData[5] = fmt.Sprintf("%v", problemBankCounterpart["stipend"])
			csvData[6] = fmt.Sprintf("%v", problemBankCounterpart["stipendforpg"])
			if len(projectAndFacilitiesCounterpart) > 0 {
				facilities := projectAndFacilitiesCounterpart["Facilities"]
				facilitiesMap := facilities.([]map[string]interface{})
				if len(facilitiesMap) > 0 {
					csvData[7] = mapToString(facilitiesMap[0])
				} else {
					csvData[7] = "Unavailable"
				}

				projects := projectAndFacilitiesCounterpart["Projects"]
				projectsMaps := projects.([]map[string]interface{})
				if len(projectsMaps) > 0 {
					csvData[8] = ""
					for i, projectMap := range projectsMaps {
						csvData[8] += fmt.Sprintln("Project: " + strconv.FormatInt(int64(i), 10))
						csvData[8] += fmt.Sprintln("Title: " + fmt.Sprintf("%v", projectMap["projectTitle"]))
						csvData[8] += fmt.Sprintln("Description: " + fmt.Sprintf("%v", projectMap["PBDescription"]))
						csvData[8] += fmt.Sprintln("Skills: " + fmt.Sprintf("%v", projectMap["SKills"]))
						csvData[8] += fmt.Sprintln("Students Required: " + fmt.Sprintf("%v", projectMap["TotalReqdStudents1"]))
						csvData[8] += fmt.Sprintln("Min CGPA: " + fmt.Sprintf("%v", projectMap["GeneralMinCGPA"]))
						csvData[8] += fmt.Sprintln("Max CGPA: " + fmt.Sprintf("%v", projectMap["GeneralMaxCGPA"]))
						csvData[8] += "\n"
					}
				}
			} else {
				csvData[7] = "Unavailable"
				csvData[8] = "Unavailable"
			}

			csvData[9] = "No"
		} else {

			temp := strings.Split(fmt.Sprintf("%v", station["Companyname"]), "-")
			var companyDomain, companyName string
			if len(temp) > 1 {
				companyDomain = strings.TrimSpace(temp[0])
				companyName = strings.TrimSpace(temp[1])
			} else {
				companyDomain = "Unavailable"
				companyName = temp[0]
			}

			csvData[0] = fmt.Sprintf("%v", station["StationId"])
			csvData[1] = companyName
			csvData[2] = fmt.Sprintf("%v", station["City"])
			csvData[3] = companyDomain
			csvData[4] = "Unavailable"
			csvData[5] = "Unavailable"
			csvData[6] = "Unavailable"
			csvData[7] = "Unavailable"
			csvData[8] = "Unavailable"
			csvData[9] = "No"
		}
		err = csvWriter.Write(csvData)
		checkErrors(err)
		fmt.Println("Written " + strconv.FormatInt(int64(i + 1), 10) + " out of " + strconv.FormatInt(length, 10))
	}
	csvWriter.Flush()
	err = csvFile.Close()
	checkErrors(err)
}

func getRequest(url string, cookies string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	checkErrors(err)

	setHeaders(req, 0, cookies, "")
	client := &http.Client{}
	resp, err := client.Do(req)
	return resp
}

func postRequest(url string, data string, cookies string, referrer string) []map[string]interface{} {
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	checkErrors(err)
	//dataContent, err := ioutil.ReadAll(data)
	checkErrors(err)
	setHeaders(req, int64(len(data)), cookies, referrer)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	//checkErrors(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalln(url + " " + strconv.FormatInt(int64(resp.StatusCode), 10))
	} else {
		fmt.Println(url + " " + strconv.FormatInt(int64(resp.StatusCode), 10))
	}
	return decodeJSON(&resp.Body)
}

func getUpdateJSON() []byte {
	csvFile, err := os.Open("StationDetails.csv")
	checkErrors(err)
	csvReader := csv.NewReader(csvFile)
	record, err := csvReader.Read()

	//Shitty method as the PSD website dev is a 10th grade kid who doesn't know that JSON libraries exist (that or he is
	//incapable of understanding how JSON works, maybe both)
	var i int64 = 1
	var jsondata = "["
	for {
		record, err = csvReader.Read()
		if err != io.EOF {
			checkErrors(err)
		} else {
			break
		}
		jsondata += "{"
		jsondata += "'isActive':'1',"
		jsondata += "'PreferenceNo':'" + strconv.FormatInt(i, 10) + "','StationId':'" + record[0] + "',"
		i += 1
		var accommodation string
		if string(record[9][0]) == "Y" || string(record[9][0]) == "y" {
			accommodation = "true"
		} else {
			accommodation = "false"
		}
		jsondata += "'Accommodation':'" + accommodation + "',"

		jsondata += "},"
	}
	jsondata = jsondata[:len(jsondata)-1]
	jsondata += "]"
	data := "{jsondata: \"" + url.QueryEscape(jsondata) + "\", jsonvalue: \"\" , contistation: \"0\"}"

	return []byte(data)

	// Good method which doesn't work as the guy doesn't use JSON but a custom shitty version of it
	// for some godforsaken reason -_-
	//
	//var finalPostData = make(map[string]interface{}, 3)
	//finalPostData["jsonvalue"] = ""
	//finalPostData["contistation"] = "0"
	//
	//var preferenceListArr []map[string]string
	//
	//record, err := csvReader.Read() //First Row has titles
	//
	////Make JSONArray for sending
	//var i int64 = 1
	//for {
	//	var StationEntry = make(map[string]string, 4)
	//	record, err = csvReader.Read()
	//	if err != io.EOF {
	//		checkErrors(err)
	//	} else {
	//		break
	//	}
	//	StationEntry["isActive"] = "1"
	//	StationEntry["PreferenceNo"] = strconv.FormatInt(i, 10)
	//	i += 1
	//	StationEntry["StationId"] = record[0]
	//	var accommodation string
	//	if string(record[7][0]) == "Y" || string(record[7][0]) == "y" {
	//		accommodation = "true"
	//	} else {
	//		accommodation = "false"
	//	}
	//	StationEntry["Accomodation"] = accommodation
	//
	//	preferenceListArr = append(preferenceListArr, StationEntry)
	//
	//}
	//
	//updateList, err := json.Marshal(preferenceListArr)
	//checkErrors(err)
	//finalPostData["jsondata"] = url.PathEscape(strings.ReplaceAll(string(updateList), "\"", "'"))
	//fmt.Println(finalPostData)
	//updateJSON, err := json.Marshal(finalPostData)
	//return updateJSON
	//fmt.Println(preferenceListArr)
}

func getStationDetails(stationId, companyId string) map[string]interface{} {

	projectTemp := make(map[string]interface{}, 2)
	referrer := "http://psd.bits-pilani.ac.in/Student/StationproblemBankDetails.aspx?CompanyId=" + companyId + "&StationId=" + stationId + "&BatchIdFor=9&PSTypeFor=2"
	getRequest(referrer, os.Args[2]) //set state variable on the shitty server, else it will return the initial or the last company you visited
	projectDetails := postRequest("http://psd.bits-pilani.ac.in/Student/StationproblemBankDetails.aspx/ViewPB", "{batchid: \"undefined\" }", os.Args[2], referrer)
	facilitiesDetails := postRequest("http://psd.bits-pilani.ac.in/Student/StationproblemBankDetails.aspx/StationFacilitiesInfo", "{StationId: \"0\"}", os.Args[2], referrer)
	projectTemp["Projects"] = projectDetails
	projectTemp["Facilities"] = facilitiesDetails

	return projectTemp
}

func main() {
	if os.Args[1] == "-g" {
		//Create CSV
		stationList := postRequest("http://psd.bits-pilani.ac.in/Student/StudentStationPreference.aspx/getinfoStation", "{CompanyId: \"0\" }", os.Args[2], "http://psd.bits-pilani.ac.in/Student/ViewActiveStationProblemBankData.aspx")
		problemBank := postRequest("http://psd.bits-pilani.ac.in/Student/ViewActiveStationProblemBankData.aspx/getPBdetail", "{batchid: \"undefined\" }", os.Args[2], "http://psd.bits-pilani.ac.in/Student/ViewActiveStationProblemBankData.aspx")

		writeCSV(stationList, problemBank)
	}
	if os.Args[1] == "-u" {
		//Update pref list on website
		updateJSON := getUpdateJSON()
		//fmt.Println(string(updateJSON))
		postRequest("http://psd.bits-pilani.ac.in/Student/StudentStationPreference.aspx/saveStudentStationPref", string(updateJSON), os.Args[2], "http://psd.bits-pilani.ac.in/Student/ViewActiveStationProblemBankData.aspx")
	}
	if os.Args[1] != "-u" && os.Args[1] != "-g" {
		fmt.Println("Wrong Argument: " + os.Args[1])
	}

}
