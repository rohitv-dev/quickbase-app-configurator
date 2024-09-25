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

	mapping := sync.Map{}

	for _, table := range tablesRes.Tables {
		fields := targetConfig.GetFields(table.ID)
		mapping.Store(table.Name, []string{})

		for _, field := range fields {
			wg.Add(1)

			go func() {
				defer wg.Done()

				if (field.FieldType == "text" || field.FieldType == "text-multi-line") && (field.Mode == "" && !field.Properties.ForeignKey) {
					log.Println(logStyle.Render("Updating Field Length for Field " + strconv.Itoa(field.ID) + "/" + field.Label + " in Table " + table.Name))
					targetConfig.UpdateFieldLength(table.ID, field.ID, field.FieldType)

					value, ok := mapping.Load(table.Name)

					if !ok {
						mapping.Store(table.Name, []string{field.Label})
					} else {
						if v, ok := value.([]string); ok {
							newValue := append(v, field.Label)
							mapping.Store(table.Name, newValue)
						}
					}
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

	mapping.Range(func(tableName, fields any) bool {
		if tableNameStr, ok := tableName.(string); ok {
			file.WriteString(tableNameStr + "\n")
			file.WriteString("--------------\n")

			if fieldsList, ok := fields.([]string); ok {
				for _, field := range fieldsList {
					file.WriteString(field + "\n")
				}
				file.WriteString("\n")
			}
		}

		return true
	})

	log.Println(boldLogStyle.Render("Updated Fields Length"))
}
