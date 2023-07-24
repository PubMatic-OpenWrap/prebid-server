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

	failed := []string{}
	data := [][]string{}
	r := bufio.NewReader(file)
	fileLine, e := Readln(r)
	i, t := 0, 0
	for e == nil {
		t += 1
		line := strings.Split(fileLine, "\t")
		if len(line) > 0 {
			oldAdunitConfigObj := new(map[string]interface{})
			if err := json.Unmarshal([]byte(line[len(line)-1]), &oldAdunitConfigObj); err != nil {
				// fmt.Println("failed-oldAdunitConfigObj: ", err, fileLine)
				data = append(data, line)
				goto l
			}

			newAdunitConfigObj := new(adunitconfig.AdUnitConfig)
			if err := json.Unmarshal([]byte(line[len(line)-1]), &newAdunitConfigObj); err != nil || newAdunitConfigObj == nil || len(newAdunitConfigObj.Config) == 0 {
				// fmt.Println("failed-newAdunitConfigObj: ", err, fileLine)
				failed = append(failed, fileLine)
				data = append(data, line)
				fmt.Println("[INVALID!!!, ", line[:len(line)-1], " current:", line[len(line)-1], "]")
				goto l
			} else {
				i += 1
			}

			newAdunitConfigBytes, err := json.Marshal(newAdunitConfigObj)
			if err != nil {
				// fmt.Println("failed-newAdunitConfigBytes: ", err, newAdunitConfigObj)
				data = append(data, line)
				goto l
			}

			newAdunitConfigObj2 := new(map[string]interface{})
			if err := json.Unmarshal(newAdunitConfigBytes, &newAdunitConfigObj2); err != nil {
				// fmt.Println("failed-newAdunitConfigObj2: ", err, newAdunitConfigObj2)
			}

			if !reflect.DeepEqual(oldAdunitConfigObj, newAdunitConfigObj2) {
				fmt.Println("[DIFFERENT!!!, ", line[:len(line)-1], " current:", line[len(line)-1], "new:", string(newAdunitConfigBytes), "]")
				assert.Equal(tst, oldAdunitConfigObj, newAdunitConfigObj2)
				data = append(data, line)
			}
		}
	l:
		fileLine, e = Readln(r)
	}
	file.Close()

	file, err = os.Open("failed_json.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, line := range failed {
		file.WriteString(line)
	}
	file.Close()

	// create a file
	file, err = os.Create("invalid_adunit_config.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// initialize csv writer
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// write all rows at once
	writer.WriteAll(data)

	fmt.Println("Total success:", i, "out of", t)
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
