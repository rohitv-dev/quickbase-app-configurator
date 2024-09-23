package main

import (
	"app-configuration/api"
	"log"
	"os"
	"strconv"
	"sync"
)

func UpdateFieldsLength(targetConfig api.Quickbase) {
	var wg sync.WaitGroup
	log.Println(boldLogStyle.Render("Updating Fields Length..."))

	tablesRes := targetConfig.GetTables()

	mapping := make(map[string][]string)

	for _, table := range tablesRes.Tables {
		fields := targetConfig.GetFields(table.ID)
		mapping[table.Name] = []string{}

		for _, field := range fields {
			wg.Add(1)

			go func() {
				defer wg.Done()

				if (field.FieldType == "text" || field.FieldType == "text-multi-line") && (field.Mode == "" && !field.Properties.ForeignKey) {
					log.Println(logStyle.Render("Updating Field Length for Field " + strconv.Itoa(field.ID) + "/" + field.Label + " in Table " + table.Name))
					targetConfig.UpdateFieldLength(table.ID, field.ID, field.FieldType)
					mapping[table.Name] = append(mapping[table.Name], field.Label)
				}
			}()
		}

		wg.Wait()
	}

	file, err := os.OpenFile("fields.txt", os.O_APPEND|os.O_CREATE, 0600)

	if err != nil {
		log.Fatal(boldErrorStyle.Render(err.Error()))
	}

	defer file.Close()

	if err := os.Truncate("fields.txt", 0); err != nil {
		log.Fatal(boldErrorStyle.Render(err.Error()))
	}

	for tableName, fields := range mapping {
		file.WriteString(tableName + "\n")
		file.WriteString("--------------\n")

		for _, field := range fields {
			file.WriteString(field + "\n")
		}
		file.WriteString("\n")
	}

	log.Println(boldLogStyle.Render("Updated Fields Length"))
}
