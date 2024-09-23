package main

import (
	"app-configuration/api"
	"app-configuration/config"
	filemanager "app-configuration/file_manager"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

func SavePages(sourceConfig api.Quickbase) {
	log.Println(boldLogStyle.Render("Processing code pages"))

	var wg sync.WaitGroup
	config := config.ReadConfig()

	for _, pageId := range config.Pages {
		wg.Add(1)
		go func() {
			defer wg.Done()

			log.Println(logStyle.Render("Saving Code Page -- " + strconv.Itoa(pageId)))

			strPageId := strconv.Itoa(pageId)
			res := sourceConfig.GetPage(strPageId)
			filemanager.SaveFile("pages/source/"+strPageId+".txt", strings.TrimSpace(res.PageBody))
		}()
	}

	wg.Wait()
}

func ReplacePages(targetConfig api.Quickbase) {
	files, err := os.ReadDir("pages/source")

	if err != nil {
		log.Fatal(errorStyle.Render(err.Error()))
	}

	var wg sync.WaitGroup

	// Looping through each file
	for _, file := range files {
		wg.Add(1)

		go func() {
			defer wg.Done()

			content := filemanager.ReadFile("pages/source/" + file.Name())

			mapping := filemanager.ReadMapping()
			flag := false

			// Replacing content
			for source, target := range mapping {
				if strings.Contains(content, source) {
					flag = true
					content = strings.ReplaceAll(content, source, target)
				}
			}

			if flag {
				pageId := strings.TrimSuffix(file.Name(), ".txt")

				log.Println(logStyle.Render("Updating Code Page -- " + pageId))

				targetConfig.ReplacePage(pageId, content)
				filemanager.SaveFile("pages/target/"+file.Name(), content)
				flag = false
			}
		}()

	}

	wg.Wait()
}
