package main

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"path/filepath"
	"strings"
	"time"
)

func doMake(arg2, arg3 string) error {
	switch arg2 {
	case "key":
		rnd := cel.RandomString(32)
		color.Yellow("32 characters encryption key: %s", rnd)

	case "migration":
		dbType := cel.DB.DataType
		if arg3 == "" {
			return errors.New("you must give the migration a name")
		}

		fileName := fmt.Sprintf("%d_%s.%s", time.Now().UnixMicro(), arg3, dbType)

		upFile := filepath.Join(cel.RootPath, "migrations", fileName+".up.sql")
		downFile := filepath.Join(cel.RootPath, "migrations", fileName+".down.sql")

		err := copyFileFromTemplate("templates/migrations/migration."+dbType+".up.sql", upFile)
		if err != nil {
			return err
		}

		err = copyFileFromTemplate("templates/migrations/migration."+dbType+".down.sql", downFile)
		if err != nil {
			return err
		}

	case "auth":
		err := doAuth()
		if err != nil {
			return err
		}

	case "handler":
		if arg3 == "" {
			return errors.New("you must give the handler a name")
		}

		fileName := filepath.Join(cel.RootPath, "handlers", strings.ToLower(arg3)+".go")

		if fileExists(fileName) {
			return errors.New(fileName + " already exists!")
		}

		data, err := templateFS.ReadFile("templates/handlers/handler.go.txt")
		if err != nil {
			return err
		}

		handler := string(data)
		handler = strings.ReplaceAll(handler, "$HANDLERNAME$", strcase.ToCamel(arg3))

		err = copyDataToFile([]byte(handler), fileName)
		if err != nil {
			return err
		}

	case "model":
		if arg3 == "" {
			return errors.New("you must give the model a name")
		}

		data, err := templateFS.ReadFile("templates/data/model.go.txt")
		if err != nil {
			return err
		}

		model := string(data)

		plur := pluralize.NewClient()

		var modelName = arg3
		var tableName = arg3

		if plur.IsPlural(arg3) {
			modelName = plur.Singular(arg3)
			tableName = strings.ToLower(tableName)
		} else {
			tableName = strings.ToLower(plur.Plural(arg3))
		}

		fileName := filepath.Join(cel.RootPath, "data", strings.ToLower(arg3)+".go")

		if fileExists(fileName) {
			return errors.New(fileName)
		}

		model = strings.ReplaceAll(model, "$MODELNAME$", strcase.ToCamel(modelName))
		model = strings.ReplaceAll(model, "$TABLENAME$", tableName)

		err = copyDataToFile([]byte(model), fileName)
		if err != nil {
			return err
		}

	case "session":
		err := doSessionTable()
		if err != nil {
			return err
		}

	case "mail":
		if arg3 == "" {
			return errors.New("you must give the mail template a name")
		}
		htmlMail := filepath.Join(cel.RootPath, "mail", strings.ToLower(arg3)+".html.tmpl")
		plainMail := filepath.Join(cel.RootPath, "mail", strings.ToLower(arg3)+".plain.tmpl")

		err := copyFileFromTemplate("templates/mailer/mail.html.tmpl", htmlMail)
		if err != nil {
			return err
		}

		err = copyFileFromTemplate("templates/mailer/mail.plain.tmpl", plainMail)
		if err != nil {
			return err
		}

	}

	return nil
}
