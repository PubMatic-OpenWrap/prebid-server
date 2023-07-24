package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// func Readln(r *bufio.Reader) (string, error) {
// 	var (
// 		isPrefix bool  = true
// 		err      error = nil
// 		line, ln []byte
// 	)
// 	for isPrefix && err == nil {
// 		line, isPrefix, err = r.ReadLine()
// 		ln = append(ln, line...)
// 	}
// 	return string(ln), err
// }

// func extractUniqueKeys(data interface{}) map[string]bool {
// 	uniqueKeys := make(map[string]bool)

// 	switch v := data.(type) {
// 	case map[string]interface{}:
// 		for key, value := range v {
// 			uniqueKeys[key] = true
// 			childKeys := extractUniqueKeys(value)
// 			for childKey := range childKeys {
// 				uniqueKeys[childKey] = true
// 			}
// 		}
// 	case []interface{}:
// 		for _, value := range v {
// 			childKeys := extractUniqueKeys(value)
// 			for childKey := range childKeys {
// 				uniqueKeys[childKey] = true
// 			}
// 		}
// 	}

// 	return uniqueKeys
// }

// func main() {
// 	Main2()
// }

func Main2() {
	file, err := os.Open("config_json_temp.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	keys := make(map[string]bool)

	r := bufio.NewReader(file)
	fileLine, e := Readln(r)
	for e == nil {
		line := strings.Split(fileLine, "\t")
		if len(line) > 0 {
			Count(line[len(line)-1], keys)
		}
		fileLine, e = Readln(r)
	}

	// Print the unique keys
	// fmt.Println("Unique keys:")
	for key := range keys {
		fmt.Println(key)
	}
}

func Count(jsonData string, keys map[string]bool) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for k, v := range data {
		if strings.ToLower(k) != "configPattern" && strings.ToLower(k) != "regex" &&
			strings.ToLower(k) != "config" && strings.ToLower(k) != "renderer" {
			keys[k] = true
		}

		if strings.ToLower(k) == "config" {
			if c, ok := v.(map[string]interface{}); ok {
				for k := range c {
					keys[k] = true
				}
			}
		}
	}

	// KEYS := listFieldNames(data)
	// fmt.Println(keys)

	// // Extract all the unique keys from the JSON data, including nested structs and maps
	// newKeys := extractUniqueKeys(data)
	// for k, v := range newKeys {
	// 	keys[k] = v
	// }
}

// func listFieldNames(v interface{}) []string {
// 	var fieldNames []string

// 	// Get the type of the struct
// 	// t := reflect.TypeOf(v)

// 	val := reflect.ValueOf(v)
// 	if val.Kind() == reflect.Ptr {
// 		val = val.Elem()
// 	}

// 	// Check if the provided value is a struct
// 	if val.Kind() == reflect.Struct {
// 		// Loop through each field
// 		for i := 0; i < val.NumField(); i++ {
// 			field := val.Type().Field(i)
// 			fieldValue := val.Field(i)

// 			fmt.Println(field.Name, field.Type.Kind())
// 			// If the field is an embedded struct (anonymous field), recursively list its fields
// 			if field.Type.Kind() == reflect.Struct {
// 				embeddedFieldNames := listFieldNames(fieldValue.Interface())
// 				fieldNames = append(fieldNames, embeddedFieldNames...)
// 			} else if field.Type.Kind() == reflect.Map { // Handle maps
// 				for _, key := range fieldValue.MapKeys() {
// 					// mapValue := val.MapIndex(key)
// 					// mapFieldNames := listFieldNames(mapValue.Interface())
// 					// fieldNames = append(fieldNames, mapFieldNames...)
// 					// fieldNames = append(fieldNames, key.String())
// 				}
// 			} else {
// 				fieldNames = append(fieldNames, field.Name)
// 			}
// 		}
// 	}

// 	return fieldNames
// }

// func deleteKeyRecursively(data interface{}, keyToDelete string) {
// 	switch v := data.(type) {
// 	case map[string]interface{}:
// 		for key := range v {
// 			if key == keyToDelete {
// 				delete(v, key)
// 			} else {
// 				deleteKeyRecursively(v[key], keyToDelete)
// 			}
// 		}
// 	case []interface{}:
// 		for i := range v {
// 			deleteKeyRecursively(v[i], keyToDelete)
// 		}
// 	}
// }

// func deleteKeyFromJSON(jsonData []byte, keyToDelete string) (string, error) {
// 	// Parse the JSON data into an interface{}
// 	var data interface{}
// 	err := json.Unmarshal(jsonData, &data)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Recursively delete the key from the JSON data
// 	deleteKeyRecursively(data, keyToDelete)

// 	// Marshal the updated data back into JSON
// 	updatedJSON, err := json.Marshal(data)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(updatedJSON), nil
// }

// func main() {
// 	file, err := os.Open("config_json_temp.txt")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	KEYS := listFieldNames(adunitconfig.AdUnitConfig{})

// 	keys :=
// 	updated := []string{}
// 	r := bufio.NewReader(file)
// 	fileLine, e := Readln(r)
// 	for e == nil {
// 		line := strings.Split(fileLine, "\t")
// 		if len(line) > 0 {
// 			updateJSON := line[len(line)-1]
// 			for _, keyToDelete := range KEYS {
// 				updateJSON_, err := deleteKeyFromJSON([]byte(updateJSON), keyToDelete)
// 				updateJSON = string(updateJSON_)
// 				if err != nil {
// 					fmt.Println(err)
// 					continue
// 				}
// 			}
// 			updated = append(updated, string(updateJSON))
// 		}
// 		fileLine, e = Readln(r)
// 	}

// 	file, err = os.Open("update_json.txt")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	for _, line := range updated {
// 		file.WriteString(line)
// 	}
// 	file.Close()
// }
