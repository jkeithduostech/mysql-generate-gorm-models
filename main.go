package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var modelTemplate = `package models

import (
    "gorm.io/gorm"
)

type {{.TableName}} struct {
{{- range .Columns }}
    {{.Name}} {{.Type}} ` + "`gorm:\"column:{{.DBName}}\"`" + `
{{- end }}
}

func ({{.TableName}}) TableName() string {
    return "{{.DBName}}"
}
`

type Column struct {
	Name   string
	Type   string
	DBName string
}

type Table struct {
	TableName string
	DBName    string
	Columns   []Column
}

func main() {
	destPath := flag.String("dest", ".", "Destination path for generated models")
	envFile := flag.String("env", ".env", "Path to .env file")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPassword := flag.String("dbpassword", "", "Database password")
	dbHost := flag.String("dbhost", "127.0.0.1", "Database host")
	dbPort := flag.String("dbport", "3306", "Database port")
	dbName := flag.String("dbname", "", "Database name")
	tables := flag.String("tables", "", "Comma-separated list of tables to generate models for")
	flag.Parse()

	// Load environment variables from .env file if it exists
	if _, err := os.Stat(*envFile); err == nil {
		err := godotenv.Load(*envFile)
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	// Override environment variables with command-line arguments if provided
	if *dbUser == "" {
		*dbUser = os.Getenv("DB_USER")
	}
	if *dbPassword == "" {
		*dbPassword = os.Getenv("DB_PASSWORD")
	}
	if *dbHost == "" {
		*dbHost = os.Getenv("DB_HOST")
	}
	if *dbPort == "" {
		*dbPort = os.Getenv("DB_PORT")
	}
	if *dbName == "" {
		*dbName = os.Getenv("DB_NAME")
	}
	if *tables == "" {
		*tables = os.Getenv("TABLES")
	}

	if *dbUser == "" || *dbPassword == "" || *dbName == "" || *tables == "" {
		log.Fatal("Database user, password, name, and tables are required")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", *dbUser, *dbPassword, *dbHost, *dbPort, *dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	tableNames := strings.Split(*tables, ",")
	for _, tableName := range tableNames {
		generateModel(db, tableName, *destPath)
	}
}

func generateModel(db *gorm.DB, tableName, destPath string) {
	var columns []Column
	_, err := db.Migrator().ColumnTypes(tableName)
	if err != nil {
		log.Fatalf("Failed to get columns for table %s: %v", tableName, err)
	}

	table := Table{
		TableName: camelCase(tableName),
		DBName:    tableName,
		Columns:   columns,
	}

	tmpl, err := template.New("model").Parse(modelTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	file, err := os.Create(fmt.Sprintf("%s/%s.go", destPath, table.TableName))
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, table)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

func camelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}
