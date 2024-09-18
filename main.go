package main

import (
	"app-configuration/api"
	"app-configuration/config"
	filemanager "app-configuration/file_manager"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
)

var (
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	logStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#008060"))
	boldErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true)
	boldLogStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#008060")).Bold(true)
)

func CreateMapping(sourceConfig api.Quickbase, targetConfig api.Quickbase) map[string]string {
	log.Println(boldLogStyle.Render("Creating mapping..."))

	mapping := make(map[string]string)

	sourceRes := sourceConfig.GetTables()
	targetRes := targetConfig.GetTables()

	filemanager.SaveJsonToFile("tables/"+sourceRes.AppId, sourceRes.Tables)
	filemanager.SaveJsonToFile("tables/"+targetRes.AppId, targetRes.Tables)

	for _, sourceTable := range sourceRes.Tables {
		for _, targetTable := range targetRes.Tables {
			if sourceTable.Name == targetTable.Name {
				mapping[sourceTable.ID] = targetTable.ID
			}
		}
	}

	mapping[sourceRes.AppId] = targetRes.AppId
	mapping[sourceConfig.UserToken] = targetConfig.UserToken
	mapping[sourceConfig.Realm] = targetConfig.Realm

	filemanager.SaveJsonToFile("mapping/mapping", mapping)

	log.Println(boldLogStyle.Render("Mapping saved"))

	return mapping
}

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

func ClearFolder(folderName string) {
	folder, err := os.ReadDir(folderName)

	if err != nil {
		log.Fatal(errorStyle.Render(err.Error()))
	}

	for _, file := range folder {
		os.RemoveAll(path.Join([]string{folderName, file.Name()}...))
	}
}

func GetQuickbaseConfigs() (api.Quickbase, api.Quickbase) {
	config := config.ReadConfig()

	sourceConfig := api.Quickbase{AppId: config.Source.Id, UserToken: config.Source.Token, Realm: config.Source.Realm}
	targetConfig := api.Quickbase{AppId: config.Target.Id, UserToken: config.Target.Token, Realm: config.Target.Realm}

	return sourceConfig, targetConfig
}

func generateAppTable() table.Model {
	sourceConfig, targetConfig := GetQuickbaseConfigs()
	config := config.ReadConfig()

	sourceApp := sourceConfig.GetApp()
	targetApp := targetConfig.GetApp()

	columns := []table.Column{
		{Title: "Type", Width: 10},
		{Title: "App ID", Width: 10},
		{Title: "App Name", Width: 30},
		{Title: "Realm", Width: 50},
		{Title: "Token", Width: 50},
	}

	rows := []table.Row{
		{"Source", config.Source.Id, sourceApp.Name, config.Source.Realm, config.Source.Token},
		{"Target", config.Target.Id, targetApp.Name, config.Target.Realm, config.Target.Token},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(2),
	)

	s := table.DefaultStyles()

	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(true)

	t.SetStyles(s)

	return t
}

func main() {
	app := &cli.App{
		Version: "v1.0.0",
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "Prints the config to console",
				Action: func(ctx *cli.Context) error {
					appTable := generateAppTable()

					fmt.Println(appTable.View())

					return nil
				},
			},
			{
				Name:  "create-config",
				Usage: "Creates config file if not present",
				Action: func(ctx *cli.Context) error {
					config.ReadConfig()

					return nil
				},
			},
			{
				Name:  "run",
				Usage: "Runs the program with both code pages and fields options",
				Action: func(ctx *cli.Context) error {
					folders := []string{"mapping", "tables", "pages/source", "pages/target", "fields/source", "fields/target"}

					for _, folder := range folders {
						ClearFolder(folder)
					}

					sourceConfig, targetConfig := GetQuickbaseConfigs()

					CreateMapping(sourceConfig, targetConfig)
					SavePages(sourceConfig)
					ReplacePages(targetConfig)
					ProcessSourceFields(sourceConfig)
					SaveFields(targetConfig)

					return nil
				},
			},
			{
				Name:  "mapping",
				Usage: "Creates the mapping from source to target",
				Action: func(ctx *cli.Context) error {
					sourceConfig, targetConfig := GetQuickbaseConfigs()

					folders := []string{"mapping", "tables"}

					for _, folder := range folders {
						ClearFolder(folder)
					}

					CreateMapping(sourceConfig, targetConfig)

					return nil
				},
			},
			{
				Name:  "pages",
				Usage: "Fetches the code pages from source app and updates them in the target app (as per the provided list in config)",
				Action: func(ctx *cli.Context) error {
					folders := []string{"mapping", "tables", "pages/source", "pages/target"}

					for _, folder := range folders {
						ClearFolder(folder)
					}

					sourceConfig, targetConfig := GetQuickbaseConfigs()

					CreateMapping(sourceConfig, targetConfig)
					SavePages(sourceConfig)
					ReplacePages(targetConfig)

					return nil
				},
			},
			{
				Name:  "fields",
				Usage: "Fetch the fields from all tables in source and updates the fields to target (if Table IDs are found)",
				Action: func(ctx *cli.Context) error {
					folders := []string{"mapping", "tables", "fields/source", "fields/target"}

					for _, folder := range folders {
						ClearFolder(folder)
					}

					sourceConfig, targetConfig := GetQuickbaseConfigs()

					CreateMapping(sourceConfig, targetConfig)
					ProcessSourceFields(sourceConfig)
					SaveFields(targetConfig)

					return nil
				},
			},
			{
				Name:  "clear",
				Usage: "Clears all the folders",
				Action: func(ctx *cli.Context) error {
					folders := []string{"mapping", "tables", "pages/source", "pages/target", "fields/source", "fields/target"}

					for _, folder := range folders {
						ClearFolder(folder)
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(errorStyle.Render(err.Error()))
	}
}
