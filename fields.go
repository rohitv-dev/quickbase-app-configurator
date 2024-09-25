package main

import (
	"app-configuration/api"
	filemanager "app-configuration/file_manager"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type TargetField struct {
	TableId   string
	TableName string
	Fields    []api.Field
}

func ProcessSourceFields(sourceConfig api.Quickbase) {
	log.Println(boldLogStyle.Render("Processing source fields"))

	var wg sync.WaitGroup
	mapping := filemanager.ReadMapping()

	// Loop through source table ids
	for tableId := range mapping {
		wg.Add(1)

		go func() {
			defer wg.Done()

			fields := sourceConfig.GetFields(tableId)
			fieldsToUpdate := make([]api.Field, 0)

			// Find only formula fields where table id exists
			for _, field := range fields {
				formula := field.Properties.Formula
				flag := false

				if len(formula) > 0 {
					for source := range mapping {
						if strings.Contains(formula, source) {
							flag = true
						}
					}
				}

				if flag {
					log.Println(logStyle.Render("Field found -- " + field.Label))
					fieldsToUpdate = append(fieldsToUpdate, field)
					flag = false
				}
			}

			if len(fieldsToUpdate) > 0 {
				filemanager.SaveJsonToFile("fields/source/"+tableId, fieldsToUpdate)
			}
		}()
	}

	wg.Wait()
}

func SaveFields(targetConfig api.Quickbase) {
	var wg sync.WaitGroup
	mapping := filemanager.ReadMapping()
	files, err := os.ReadDir("fields/source")

	if err != nil {
		log.Fatal(errorStyle.Render(err.Error()))
	}

	for _, file := range files {
		fields := filemanager.ReadFields("fields/source/" + file.Name())
		targetTable := mapping[strings.TrimSuffix(file.Name(), ".json")]

		for _, field := range fields {
			wg.Add(1)

			go func() {
				defer wg.Done()
				formula := field.Properties.Formula

				for source, target := range mapping {
					formula = strings.ReplaceAll(formula, source, target)
				}

				field.Properties.Formula = formula

				log.Println(logStyle.Render("Updating Field -- " + field.Label))

				targetConfig.UpdateField(targetTable, strconv.Itoa(field.ID), formula)
				filemanager.SaveJsonToFile("fields/target/"+targetTable+"_"+strconv.Itoa(field.ID)+"_"+field.Label, field)
			}()
		}
	}

	wg.Wait()
}

func SaveTargetFields(targetConfig api.Quickbase) {
	var wg sync.WaitGroup

	log.Println(boldLogStyle.Render("Saving Target Fields..."))

	if _, err := os.Stat("target_fields.json"); err == nil {
		log.Println(warningStyle.Render("target_fields.json already exists, using it"))
		return
	}

	tablesRes := targetConfig.GetTables()

	targetFields := make([]TargetField, len(tablesRes.Tables))

	for index, table := range tablesRes.Tables {
		wg.Add(1)

		go func(i int, t api.Table) {
			defer wg.Done()

			fields := targetConfig.GetFields(t.ID)

			targetFields[index] = TargetField{
				TableId:   t.ID,
				TableName: filemanager.SanitizeFileName(t.Name),
				Fields:    fields,
			}

		}(index, table)
	}

	wg.Wait()

	filemanager.SaveJsonToFile("target_fields", targetFields)
}

func GetTextFields() []TargetField {
	targetFields := filemanager.ReadJSONFile[[]TargetField]("target_fields.json")

	for index, target := range targetFields {
		textFields := make([]api.Field, 0)

		for _, field := range target.Fields {
			if (field.FieldType == "text" || field.FieldType == "text-multi-line") && (field.Mode == "" && !field.Properties.ForeignKey) {
				textFields = append(textFields, field)
			}
		}

		targetFields[index].Fields = textFields
	}

	return targetFields
}

func GetToUpdateTextFields() []TargetField {
	targetFields := GetTextFields()

	for index, target := range targetFields {
		textFields := make([]api.Field, 0)

		for _, field := range target.Fields {
			if (field.FieldType == "text" && field.Properties.MaxLength != 50) || (field.FieldType == "text-multi-line" && field.Properties.MaxLength != 200) {
				textFields = append(textFields, field)
			}
		}

		targetFields[index].Fields = textFields
	}

	return targetFields
}

func GetFileFields() []TargetField {
	targetFields := filemanager.ReadJSONFile[[]TargetField]("target_fields.json")

	for index, target := range targetFields {
		fileFields := make([]api.Field, 0)

		for _, field := range target.Fields {
			if field.FieldType == "file" {
				fileFields = append(fileFields, field)
			}
		}

		targetFields[index].Fields = fileFields
	}

	return targetFields
}
