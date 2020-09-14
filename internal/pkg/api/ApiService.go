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

package api

import (
	"encoding/json"
	"github.com/wso2/micro-gw/internal/pkg/models"
	"github.com/wso2/micro-gw/internal/pkg/mysqlDB"
	"github.com/wso2/micro-gw/internal/pkg/nodeUpdate"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

//This package contains the REST API for the control plane configurations

type RESTService struct{}

// TODO: Implement. Simply copy the swagger content to the location defined in the config or directly deploy the api.
// Deploy API in microgateway.
func (rest *RESTService) ApiPOST(w http.ResponseWriter, r *http.Request) {
	labels := r.FormValue("labels")
	apiName := r.FormValue("apiName")
	uploadedFile, _, err := r.FormFile("swaggerFile")
	if err != nil {
		log.Println(err)
	}
	defer uploadedFile.Close()
	fileContent, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		log.Println(err)
	}

	var _ bool
	var response models.APICtlResponse

	labelArray := strings.Split(labels, ",")
	for _, label := range labelArray {
		_, err = mysqlDB.AddSwaggerFileToDB(label, apiName, fileContent)
		swaggerFiles, err := mysqlDB.GetSwaggerFilesFromDBForLabel(label)
		if err != nil {
			log.Println(err)
		} else {
			nodeUpdate.UpdateEnvoyMgw(swaggerFiles)
		}

		if err != nil {
			log.Println("Failed to add API", err)
			response.Message = "Failed to add API. " + err.Error()
		} else {
			log.Println("Your API is added")
			response.Message = "Your API is added"
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
	}
}

// Update deployed api
func (rest *RESTService) ApiPUT(w http.ResponseWriter, r *http.Request) {
	labels := r.FormValue("labels")
	apiName := r.FormValue("apiName")
	uploadedFile, _, err := r.FormFile("swaggerFile")
	if err != nil {
		log.Fatal(err)
	}
	defer uploadedFile.Close()
	fileContent, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		log.Fatal(err)
	}

	var _ bool
	var response models.APICtlResponse

	labelArray := strings.Split(labels, ",")
	for _, label := range labelArray {
		_, err = mysqlDB.UpdateSwaggerFileDB(label, apiName, fileContent)
		swaggerFiles, err := mysqlDB.GetSwaggerFilesFromDBForLabel(label)
		if err != nil {
			log.Println(err)
		} else {
			nodeUpdate.UpdateEnvoyMgw(swaggerFiles)
		}

		if err != nil {
			log.Println("Failed to update API. ", err)
			response.Message = "Failed to update API. " + err.Error()
		} else {
			log.Println("Your API is updated")
			response.Message = "Your API is updated"
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
	}
}

// Remove a deployed api
func (rest *RESTService) ApiDELETE(w http.ResponseWriter, r *http.Request) {
	labels := r.FormValue("labels")
	apiName := r.FormValue("apiName")

	var response models.APICtlResponse
	labelArray := strings.Split(labels, ",")
	for _, label := range labelArray {
		_, err := mysqlDB.DeleteSwaggerFileFromDB(label, apiName)

		swaggerFiles, err := mysqlDB.GetSwaggerFilesFromDBForLabel(label)
		if err != nil {
			log.Println(err)
		} else {
			nodeUpdate.UpdateEnvoyMgw(swaggerFiles)
		}

		if err != nil {
			log.Println("Failed to delete API. ", err)
			response.Message = "Failed to delete API. " + err.Error()
		} else {
			log.Println("Your API is deleted")
			response.Message = "Your API is deleted"
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
	}
}
