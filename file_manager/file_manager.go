package filemanager

import (
	"app-configuration/api"
	"encoding/json"
	"io"
	"log"
	"os"
	"regexp"
)

func SanitizeFileName(fileName string) string {
	regex, err := regexp.Compile(`[?!&*/]`)

	if err != nil {
		log.Fatal(err)
	}

	return regex.ReplaceAllString(fileName, "_")
}

func ReadMapping() map[string]string {
	file, err := os.Open("mapping/mapping.json")

	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)

	var data map[string]string

	decoder.Decode(&data)

	return data
}

func ReadJSONFile[T any](filePath string) T {
	var result T

	file, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)

	decoder.Decode(&result)

	return result
}

func SaveJsonToFile(fileName string, content any) error {
	file, err := os.Create(SanitizeFileName(fileName) + ".json")

	if err != nil {
		return nil
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent(" ", "  ")

	return encoder.Encode(content)
}

func ReadFile(fileName string) string {
	content, err := os.ReadFile(fileName)

	if err != nil {
		log.Fatal(err)
	}

	return string(content)
}

func ReadTextFile(filePath string) string {
	file, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)

	if err != nil {
		log.Fatal(err)
	}

	content := string(bytes)

	return content
}

func SaveFile(fileName string, content string) {
	err := os.WriteFile(SanitizeFileName(fileName), []byte(content), 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func ReadFields(fileName string) []api.Field {
	file, err := os.ReadFile(fileName)

	if err != nil {
		log.Fatal(err)
	}

	var fields []api.Field

	json.Unmarshal(file, &fields)

	return fields
}
