package main

import (
	filemanager "app-configuration/file_manager"
	"log"
	"os"
	"strings"
)

func CustomRules() {
	log.Println(boldLogStyle.Render("Processing Custom Text Rules..."))

	content := filemanager.ReadTextFile("placeholders/custom_text.txt")
	header := filemanager.ReadTextFile("placeholders/custom_text_header.txt")

	fileContent := filemanager.ReadTextFile("placeholders/custom_file.txt")

	ClearFolder("rules")

	targetFields := GetTextFields()
	fileFields := GetFileFields()

	for _, target := range targetFields {
		fieldsList := ""
		fileName := "rules/" + target.TableName + ".txt"

		if len(target.Fields) == 0 {
			continue
		}

		file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0600)

		if err != nil {
			log.Fatal(boldErrorStyle.Render(err.Error()))
		}

		defer file.Close()

		if err := os.Truncate(fileName, 0); err != nil {
			log.Fatal(boldErrorStyle.Render(err.Error()))
		}

		file.WriteString(header)
		file.WriteString("\n\n")

		for _, field := range target.Fields {
			fieldName := strings.ReplaceAll(field.Label, " ", "")

			newContent := strings.ReplaceAll(content, "[ABCDName]", "["+field.Label+"]")
			newContent = strings.ReplaceAll(newContent, "ABCDVarName", fieldName)

			file.WriteString(newContent)
			file.WriteString("\n\n")

			fieldsList += "$" + fieldName + ", "
		}

		file.WriteString(fieldsList)
		file.WriteString("\n\n")
	}

	for _, target := range fileFields {
		fileName := "rules/" + target.TableName + ".txt"

		if len(target.Fields) == 0 {
			continue
		}

		file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0600)

		if err != nil {
			log.Fatal(boldErrorStyle.Render(err.Error()))
		}

		defer file.Close()

		for _, field := range target.Fields {
			fieldName := strings.ReplaceAll(field.Label, " ", "")

			newFileContent := strings.ReplaceAll(fileContent, "[ABCDName]", "["+field.Label+"]")
			newFileContent = strings.ReplaceAll(newFileContent, "ABCDVarName", fieldName)

			file.WriteString(newFileContent)
			file.WriteString("\n\n")
		}
	}
}
