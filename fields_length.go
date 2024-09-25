package main

import (
	"app-configuration/api"
	"log"
	"os"
	"sync"
)

func UpdateFieldsLength(targetConfig api.Quickbase) {
	var wg sync.WaitGroup

	textFields := GetToUpdateTextFields()

	file, err := os.OpenFile("fields.txt", os.O_APPEND|os.O_CREATE, 0600)

	if err != nil {
		log.Fatal(boldErrorStyle.Render(err.Error()))
	}

	defer file.Close()

	if err := os.Truncate("fields.txt", 0); err != nil {
		log.Fatal(boldErrorStyle.Render(err.Error()))
	}

	for _, target := range textFields {
		if len(target.Fields) == 0 {
			continue
		}

		for _, field := range target.Fields {
			wg.Add(1)

			go func() {
				defer wg.Done()

				targetConfig.UpdateFieldLength(target.TableId, field.ID, field.FieldType)
			}()
		}
	}

	wg.Wait()

	for _, target := range textFields {
		file.WriteString(target.TableName + "\n")
		file.WriteString("--------------\n")

		for _, field := range target.Fields {
			file.WriteString(field.Label + "\n")
		}
		file.WriteString("\n")
	}

	log.Println(boldLogStyle.Render("Updated Fields Length"))
}
