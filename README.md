# MysqL GORM Model Generator

This Go application generates GORM models based on existing tables in a MySQL database. It takes in arguments for the destination path, database connection details, and a list of tables to build models for. Additionally, it can read the database connection information from a `.env` file.

## Features

- Generate GORM models for specified tables in a MySQL database.
- Command-line arguments for database connection details and destination path.
- Support for loading database connection details from a `.env` file.

## Usage

### Command-Line Arguments

The application accepts the following command-line arguments:

- `-dest`: Destination path for generated models (default: `.`).
- `-env`: Path to `.env` file (default: `.env`).
- `-dbuser`: Database user.
- `-dbpassword`: Database password.
- `-dbhost`: Database host (default: `127.0.0.1`).
- `-dbport`: Database port (default: `3306`).
- `-dbname`: Database name.
- `-tables`: Comma-separated list of tables to generate models for.

### Example Command

```sh
go run main.go -dest=./models -dbuser=user -dbpassword=password -dbhost=127.0.0.1 -dbport=3306 -dbname=dbname -tables="table1,table2"
```

### Example Command With .env

```sh
go run main.go -dest=./models -env=.env
```

### Example Command With .env with overrideing values

```sh
go run main.go -dest=./models -env=.env -tables="table1,table2"
```
