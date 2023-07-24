package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func main() { Main(&testing.T{}) }

func Main(tst *testing.T) {
	file, err := os.Open("config_json_temp.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	invalidAdunitsMap := make(map[string]bool)
	validAdunitsMap := make(map[string]bool)
	// failed := []string{}
	invalidAUC := [][]string{}
	validAUC := [][]string{}
	r := bufio.NewReader(file)
	fileLine, e := Readln(r)
	v, iv, t := 0, 0, 0
	for e == nil {
		t += 1
		line := strings.Split(fileLine, "\t")
		if len(line) > 0 {
			oldAdunitConfigObj := new(map[string]interface{})
			if err := json.Unmarshal([]byte(line[len(line)-1]), &oldAdunitConfigObj); err != nil {
				// fmt.Println("failed-oldAdunitConfigObj: ", err, fileLine)
				Count(line[len(line)-1], invalidAdunitsMap)
				invalidAUC = append(invalidAUC, line)
				iv += 1
				goto l
			}

			newAdunitConfigObj := new(adunitconfig.AdUnitConfig)
			if err := json.Unmarshal([]byte(line[len(line)-1]), &newAdunitConfigObj); err != nil || newAdunitConfigObj == nil || len(newAdunitConfigObj.Config) == 0 {
				// fmt.Println("failed-newAdunitConfigObj: ", err, fileLine)
				// failed = append(failed, fileLine)
				invalidAUC = append(invalidAUC, line)
				Count(line[len(line)-1], invalidAdunitsMap)
				iv += 1
				// fmt.Println("[INVALID!!!, ", line[:len(line)-1], " current:", line[len(line)-1], "]")
				goto l
			}

			newAdunitConfigBytes, err := json.Marshal(newAdunitConfigObj)
			if err != nil {
				// fmt.Println("failed-newAdunitConfigBytes: ", err, newAdunitConfigObj)
				invalidAUC = append(invalidAUC, line)
				Count(line[len(line)-1], invalidAdunitsMap)
				iv += 1
				goto l
			}

			newAdunitConfigObj2 := new(map[string]interface{})
			if err := json.Unmarshal(newAdunitConfigBytes, &newAdunitConfigObj2); err != nil {
				// fmt.Println("failed-newAdunitConfigObj2: ", err, newAdunitConfigObj2)
			}

			if !reflect.DeepEqual(oldAdunitConfigObj, newAdunitConfigObj2) {
				// fmt.Println("[DIFFERENT!!!, ", line[:len(line)-1], " current:", line[len(line)-1], "new:", string(newAdunitConfigBytes), "]")
				assert.Equal(tst, oldAdunitConfigObj, newAdunitConfigObj2)
				invalidAUC = append(invalidAUC, line)
				Count(line[len(line)-1], invalidAdunitsMap)
				iv += 1
			} else {
				Count(line[len(line)-1], validAdunitsMap)
				validAUC = append(validAUC, line)
				v += 1
			}
		}
	l:
		fileLine, e = Readln(r)
	}
	file.Close()

	// create a file
	file, err = os.Create("invalid_adunits.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// initialize csv writer
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// write all rows at once
	writer.WriteAll(invalidAUC)

	// create a file
	file, err = os.Create("valid_adunits.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// initialize csv writer
	writer = csv.NewWriter(file)
	defer writer.Flush()
	// write all rows at once
	writer.WriteAll(validAUC)

	writeMapKeysToFile("valid_adunits_list.txt", validAdunitsMap)
	writeMapKeysToFile("invalid_adunits_list.txt", invalidAdunitsMap)

	fmt.Println("LIVE PROFILE: Total:", t, ", Valid:", v, "Invalid:", iv, "Invalid Percentage: ", fmt.Sprintf("%.0f%%", (float64(iv)/float64(t))*100.0))

	vl := len(validAdunitsMap)
	ivl := len(invalidAdunitsMap)
	tl := vl + ivl
	fmt.Println("ADUNITS: Total:", +tl,
		", Valid:", vl,
		"Invalid:", ivl,
		"Invalid Percentage: ", fmt.Sprintf("%.0f%%", (float64(ivl)/float64(tl))*100.0))
}

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func writeMapKeysToFile(filename string, dataMap map[string]bool) error {
	// Open the file in write-only mode, create if it doesn't exist, and append if it exists
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the keys of the map to the file line by line
	for key := range dataMap {
		_, err := fmt.Fprintln(file, key)
		if err != nil {
			return err
		}
	}

	return nil
}
