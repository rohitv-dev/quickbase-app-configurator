package main

import (
	"app-configuration/api"
	"app-configuration/config"
	filemanager "app-configuration/file_manager"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
)

var (
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	logStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#008060"))
	warningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))
	boldErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true)
	boldLogStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#008060")).Bold(true)

	folders = []string{"pages", "pages/source", "pages/target", "fields", "fields/source", "fields/target", "tables", "mapping", "rules", "placeholders"}
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

	if sourceConfig.UserToken != targetConfig.UserToken {
		mapping[sourceConfig.UserToken] = targetConfig.UserToken
	}

	if sourceConfig.Realm != targetConfig.Realm {
		mapping[sourceConfig.Realm] = targetConfig.Realm
	}

	filemanager.SaveJsonToFile("mapping/mapping", mapping)

	log.Println(boldLogStyle.Render("Mapping saved"))

	return mapping
}

func VerifyFolders() {
	for _, folder := range folders {
		if _, err := os.Stat(folder); err != nil {
			err = os.Mkdir(folder, 0755)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			if folder != "placeholders" {
				ClearFolder(folder)
			}
		}
	}
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
		{Title: "App Name", Width: 70},
		{Title: "Realm", Width: 70},
		{Title: "Token", Width: 70},
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
					VerifyFolders()

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
				Name:  "fieldslength",
				Usage: "Updates the maximum length of text and multiline fields",
				Action: func(ctx *cli.Context) error {
					VerifyFolders()

					_, targetConfig := GetQuickbaseConfigs()

					SaveTargetFields(targetConfig)
					UpdateFieldsLength(targetConfig)
					VerifyFieldsLength(targetConfig)

					return nil
				},
			},
			{
				Name:  "verifyfields",
				Usage: "Verify field max length for all fields",
				Action: func(ctx *cli.Context) error {
					VerifyFolders()

					_, targetConfig := GetQuickbaseConfigs()

					VerifyFieldsLength(targetConfig)

					return nil
				},
			},
			{
				Name:  "rules",
				Usage: "Generates text and file rules to include in custom data rules",
				Action: func(ctx *cli.Context) error {
					VerifyFolders()

					_, targetConfig := GetQuickbaseConfigs()

					SaveTargetFields(targetConfig)
					CustomRules()

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
				Name:  "verify",
				Usage: "Creates folders if not present, or clears files of existing folders",
				Action: func(ctx *cli.Context) error {
					VerifyFolders()

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(errorStyle.Render(err.Error()))
	}
}
