/*
 *  Copyright (c) 2020, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package mysqlDB

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wso2/micro-gw/internal/pkg/models"
	"log"
)

var (
	//ctx context.Context
	db *sql.DB
)

func GetAllSwaggerFilesFromDB() ([]models.SwaggerFile, error) {
	var swaggerFiles []models.SwaggerFile
	db, err := connectToMySQLDB()
	if db != nil {
		defer db.Close()
		if err != nil {
			log.Println(err)
		} else {
			// perform a mysqlDB.Query select
			selectQuery, err := db.Prepare("SELECT * FROM swagger_file;")
			if err != nil {
				log.Println(err)
			}
			defer selectQuery.Close()

			var fileId int
			var fileLabel string
			var apiName string
			var file []byte

			rows, err := selectQuery.Query()
			if err != nil {
				log.Println(err)
			}
			defer rows.Close()
			for rows.Next() {
				err = rows.Scan(&fileId, &fileLabel, &apiName, &file)
				if err != nil {
					log.Println(err)
				}

				//log.Println(fileId, fileLabel, apiName, file)
				swaggerFiles = append(swaggerFiles, models.SwaggerFile{FileID: fileId, Label: fileLabel, ApiName: apiName, File: file})
			}
			if err := rows.Err(); err != nil {
				log.Println(err)
			}
			switch {
			case err == sql.ErrNoRows:
				log.Println("No swagger detected")
			case err != nil:
				log.Println(err)
			default:
				log.Println("Swagger files found")
			}
		}
	}
	return swaggerFiles, err
}

func GetSwaggerFilesFromDBForLabel(label string) ([]models.SwaggerFile, error) {
	var swaggerFiles []models.SwaggerFile
	db, err := connectToMySQLDB()
	if db != nil {
		defer db.Close()
		if err != nil {
			log.Println(err)
		} else {
			// perform a mysqlDB.Query select
			selectQuery, err := db.Prepare("SELECT * FROM swagger_file WHERE label=?;")
			if err != nil {
				log.Println(err)
			}
			defer selectQuery.Close()

			var fileId int
			var fileLabel string
			var apiName string
			var file []byte

			rows, err := selectQuery.Query(label)
			if err != nil {
				log.Println(err)
			}
			defer rows.Close()

			for rows.Next() {
				err := rows.Scan(&fileId, &fileLabel, &apiName, &file)
				if err != nil {
					log.Println(err)
				}

				//log.Println(fileId, fileLabel, apiName, file)
				swaggerFiles = append(swaggerFiles, models.SwaggerFile{FileID: fileId, Label: fileLabel, ApiName: apiName, File: file})
			}
			if err := rows.Err(); err != nil {
				log.Println(err)
			}
			switch {
			case err == sql.ErrNoRows:
				log.Printf("No swagger files detected under label %s\n", label)
			case err != nil:
				log.Println(err)
			default:
				log.Printf("Swagger files with label %s found\n", label)
			}
		}
	}
	return swaggerFiles, err
}

func GetSwaggerFileFromDB(label string, apiName string) (models.SwaggerFile, error) {
	swaggerFile := models.SwaggerFile{}
	db, err := connectToMySQLDB()
	if db != nil {
		defer db.Close()
		if err != nil {
			log.Println(err)
		} else {
			// perform a mysqlDB.Query select
			selectQuery, err := db.Prepare("SELECT * FROM swagger_file WHERE label=? AND api_name=?;")
			if err != nil {
				log.Println(err)
			}
			defer selectQuery.Close()

			err = selectQuery.QueryRow(label, apiName).Scan(&swaggerFile)

			switch {
			case err == sql.ErrNoRows:
				log.Printf("No swagger file with label %s and file Name %s\n", label, apiName)
			case err != nil:
				log.Fatal(err)
			default:
				log.Printf("Swagger file with label %s and file Name %s\n", label, apiName)
			}
		}
	}
	return swaggerFile, err
}

func AddSwaggerFileToDB(label string, apiName string, file []byte) (bool, error) {
	db, err := connectToMySQLDB()
	if db != nil {
		defer db.Close()
		if err != nil {
			log.Println(err)
		} else {
			// perform a mysqlDB.Query insert
			insertQuery, err := db.Prepare("INSERT INTO swagger_file (label, api_name, file) VALUES (?, ?, ?);")
			_, err = insertQuery.Exec(label, apiName, file)

			// if there is an error inserting, handle it
			if err != nil {
				log.Println(err)
			}
			// be careful deferring Queries if you are using transactions
			defer insertQuery.Close()
		}
	}
	return true, err
}

func UpdateSwaggerFileDB(label string, apiName string, file []byte) (bool, error) {
	db, err := connectToMySQLDB()
	if db != nil {
		defer db.Close()
		if err != nil {
			log.Println(err)
		} else {
			// perform a mysqlDB.Query insert
			updateQuery, err := db.Prepare("UPDATE swagger_file SET file = ? WHERE label = ? AND api_name = ?;")
			_, err = updateQuery.Exec(file, label, apiName)

			// if there is an error inserting, handle it
			if err != nil {
				log.Println(err)
			}
			// be careful deferring Queries if you are using transactions
			defer updateQuery.Close()
		}
	}
	return true, err
}

func DeleteSwaggerFileFromDB(label string, apiName string) (bool, error) {
	db, err := connectToMySQLDB()
	if db != nil {
		defer db.Close()
		if err != nil {
			log.Println(err)
		} else {
			// perform a mysqlDB.Query insert
			deleteQuery, err := db.Prepare("DELETE FROM swagger_file WHERE label = ? AND api_name = ?;")
			_, err = deleteQuery.Exec(label, apiName)

			// if there is an error inserting, handle it
			if err != nil {
				log.Println(err)
			}
			// be careful deferring Queries if you are using transactions
			defer deleteQuery.Close()
		}
	}
	return true, err
}

func selectMgwDB(db *sql.DB) (*sql.DB, error) {
	_, err := db.Exec("USE mgwDB;")
	if err != nil {
		log.Println(err)
		_, err = db.Exec("CREATE DATABASE mgwDB;")
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Database created successfully..")
		}
	} else {
		log.Println("DB selected successfully..")
	}
	return db, err
}

func selectSwaggerFileTable(db *sql.DB) (*sql.DB, error) {
	stmt, err := db.Prepare("CREATE Table swagger_file (file_id int NOT NULL AUTO_INCREMENT, label varchar(255), api_name varchar(255), file BLOB NOT NULL, PRIMARY KEY (file_id));")
	if err != nil {
		log.Println(err)
		_, err = stmt.Exec()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Table swagger_file created successfully..")
		}
	}
	return db, err
}

func connectToMySQLDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "<datasource-name>")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("DB Connection successful")
	}
	db, err = selectMgwDB(db)
	if err != nil {
		log.Println(err)
	} else {
		db, err = selectSwaggerFileTable(db)
		if err != nil {
			log.Println(err)
		}
	}
	return db, err
}
