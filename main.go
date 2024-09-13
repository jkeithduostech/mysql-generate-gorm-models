package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/jinzhu/inflection"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var modelTemplate = `package models

{{if .ModelImports}}
import (
{{range .ModelImports}}
	"{{.}}"
{{end}}
)
{{end}}    


type {{.TableName}} struct {
{{- range .Columns }}
    {{.Name}} {{.Type}} ` + "`gorm:\"column:{{.GormName}}\"`" + `
{{- end }}
}

func ({{.TableName}}) TableName() string {
    return "{{.DBTableName}}"
}
`

type Column struct {
	Name     string
	GormName string
	Type     string
}

type Table struct {
	TableName    string
	DBTableName  string
	Columns      []Column
	ModelImports []string
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
		log.Printf("Loaded environment variables from %s", *envFile)

		// Map of flag pointers to their corresponding environment variable names
		envVars := map[string]*string{
			"DESTPATH":    destPath,
			"DB_USER":     dbUser,
			"DB_PASSWORD": dbPassword,
			"DB_HOST":     dbHost,
			"DB_PORT":     dbPort,
			"DB_NAME":     dbName,
			"TABLES":      tables,
		}

		// Override flag variables with environment variables if they exist
		for envVar, flagPtr := range envVars {
			if val, exists := os.LookupEnv(envVar); exists {
				*flagPtr = val
			}
		}
	}

	// Override environment variables with command-line arguments if provided
	if *destPath == "." {
		*destPath = os.Getenv("DESTPATH")
	}
	if *dbUser == "" {
		*dbUser = os.Getenv("DB_USER")
	}
	if *dbPassword == "" {
		*dbPassword = os.Getenv("DB_PASSWORD")
	}
	if *dbHost == "127.0.0.1" {
		*dbHost = os.Getenv("DB_HOST")
	}
	if *dbPort == "3306" {
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
	log.Printf("Connecting to database: %s", dsn)
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
	var modelImports []string
	columnTypes, err := db.Migrator().ColumnTypes(tableName)
	if err != nil {
		log.Fatalf("Failed to get columns for table %s: %v", tableName, err)
	}

	for _, columnType := range columnTypes {
		modelColumnType := columnType.DatabaseTypeName()
		// Add special handling for datetime columns
		switch columnType.DatabaseTypeName() {
		case "datetime", "timestamp":
			modelColumnType = "time.Time"
			if !strings.Contains(strings.Join(modelImports, ","), "time") {
				modelImports = append(modelImports, "time")
			}
		case "tinyint":
			modelColumnType = "int"
		case "varchar":
			modelColumnType = "string"
		}

		column := Column{
			Name:     camelCase(columnType.Name()),
			Type:     modelColumnType,
			GormName: columnType.Name(),
			// Add other fields as necessary
		}
		columns = append(columns, column)
	}

	// depluralize table name
	depluraizedTableName := inflection.Singular(tableName)

	table := Table{
		TableName:    camelCase(depluraizedTableName),
		Columns:      columns,
		DBTableName:  tableName,
		ModelImports: modelImports,
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
	caser := cases.Title(language.Und)
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = caser.String(parts[i])
	}
	return strings.Join(parts, "")
}
