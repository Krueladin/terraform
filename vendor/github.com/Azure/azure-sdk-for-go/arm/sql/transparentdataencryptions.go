package sql

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"net/http"
)

// TransparentDataEncryptionsClient is the the Azure SQL Database management API provides a RESTful set of web services
// that interact with Azure SQL Database services to manage your databases. The API enables you to create, retrieve,
// update, and delete databases.
type TransparentDataEncryptionsClient struct {
	ManagementClient
}

// NewTransparentDataEncryptionsClient creates an instance of the TransparentDataEncryptionsClient client.
func NewTransparentDataEncryptionsClient(subscriptionID string) TransparentDataEncryptionsClient {
	return NewTransparentDataEncryptionsClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewTransparentDataEncryptionsClientWithBaseURI creates an instance of the TransparentDataEncryptionsClient client.
func NewTransparentDataEncryptionsClientWithBaseURI(baseURI string, subscriptionID string) TransparentDataEncryptionsClient {
	return TransparentDataEncryptionsClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// CreateOrUpdate creates or updates a database's transparent data encryption configuration.
//
// resourceGroupName is the name of the resource group that contains the resource. You can obtain this value from the
// Azure Resource Manager API or the portal. serverName is the name of the server. databaseName is the name of the
// database for which setting the transparent data encryption applies. transparentDataEncryptionName is the name of the
// transparent data encryption configuration. parameters is the required parameters for creating or updating
// transparent data encryption.
func (client TransparentDataEncryptionsClient) CreateOrUpdate(resourceGroupName string, serverName string, databaseName string, transparentDataEncryptionName string, parameters TransparentDataEncryption) (result TransparentDataEncryption, err error) {
	req, err := client.CreateOrUpdatePreparer(resourceGroupName, serverName, databaseName, transparentDataEncryptionName, parameters)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.TransparentDataEncryptionsClient", "CreateOrUpdate", nil, "Failure preparing request")
		return
	}

	resp, err := client.CreateOrUpdateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "sql.TransparentDataEncryptionsClient", "CreateOrUpdate", resp, "Failure sending request")
		return
	}

	result, err = client.CreateOrUpdateResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.TransparentDataEncryptionsClient", "CreateOrUpdate", resp, "Failure responding to request")
	}

	return
}

// CreateOrUpdatePreparer prepares the CreateOrUpdate request.
func (client TransparentDataEncryptionsClient) CreateOrUpdatePreparer(resourceGroupName string, serverName string, databaseName string, transparentDataEncryptionName string, parameters TransparentDataEncryption) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"databaseName":                  autorest.Encode("path", databaseName),
		"resourceGroupName":             autorest.Encode("path", resourceGroupName),
		"serverName":                    autorest.Encode("path", serverName),
		"subscriptionId":                autorest.Encode("path", client.SubscriptionID),
		"transparentDataEncryptionName": autorest.Encode("path", transparentDataEncryptionName),
	}

	const APIVersion = "2014-04-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsJSON(),
		autorest.AsPut(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Sql/servers/{serverName}/databases/{databaseName}/transparentDataEncryption/{transparentDataEncryptionName}", pathParameters),
		autorest.WithJSON(parameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// CreateOrUpdateSender sends the CreateOrUpdate request. The method will close the
// http.Response Body if it receives an error.
func (client TransparentDataEncryptionsClient) CreateOrUpdateSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// CreateOrUpdateResponder handles the response to the CreateOrUpdate request. The method always
// closes the http.Response Body.
func (client TransparentDataEncryptionsClient) CreateOrUpdateResponder(resp *http.Response) (result TransparentDataEncryption, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// Get gets a database's transparent data encryption configuration.
//
// resourceGroupName is the name of the resource group that contains the resource. You can obtain this value from the
// Azure Resource Manager API or the portal. serverName is the name of the server. databaseName is the name of the
// database for which the transparent data encryption applies. transparentDataEncryptionName is the name of the
// transparent data encryption configuration.
func (client TransparentDataEncryptionsClient) Get(resourceGroupName string, serverName string, databaseName string, transparentDataEncryptionName string) (result TransparentDataEncryption, err error) {
	req, err := client.GetPreparer(resourceGroupName, serverName, databaseName, transparentDataEncryptionName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.TransparentDataEncryptionsClient", "Get", nil, "Failure preparing request")
		return
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "sql.TransparentDataEncryptionsClient", "Get", resp, "Failure sending request")
		return
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.TransparentDataEncryptionsClient", "Get", resp, "Failure responding to request")
	}

	return
}

// GetPreparer prepares the Get request.
func (client TransparentDataEncryptionsClient) GetPreparer(resourceGroupName string, serverName string, databaseName string, transparentDataEncryptionName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"databaseName":                  autorest.Encode("path", databaseName),
		"resourceGroupName":             autorest.Encode("path", resourceGroupName),
		"serverName":                    autorest.Encode("path", serverName),
		"subscriptionId":                autorest.Encode("path", client.SubscriptionID),
		"transparentDataEncryptionName": autorest.Encode("path", transparentDataEncryptionName),
	}

	const APIVersion = "2014-04-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Sql/servers/{serverName}/databases/{databaseName}/transparentDataEncryption/{transparentDataEncryptionName}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client TransparentDataEncryptionsClient) GetSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client TransparentDataEncryptionsClient) GetResponder(resp *http.Response) (result TransparentDataEncryption, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
