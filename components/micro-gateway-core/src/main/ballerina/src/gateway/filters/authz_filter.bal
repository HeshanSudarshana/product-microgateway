// Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file   except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import ballerina/cache;
import ballerina/http;
import ballerina/runtime;

// authorization filter which wraps the ballerina in built authorization filter.

public type OAuthzFilter object {

    public http:AuthzFilter authzFilter;

    public function __init(cache:Cache positiveAuthzCache, cache:Cache negativeAuthzCache, string[][]? scopes) {
        http:AuthzHandler authzHandler = new (positiveAuthzCache, negativeAuthzCache);
        self.authzFilter = new (authzHandler, scopes);
    }

    public function filterRequest(http:Caller caller, http:Request request, http:FilterContext context) returns boolean {
        if (context.attributes.hasKey(SKIP_ALL_FILTERS) && <boolean>context.attributes[SKIP_ALL_FILTERS]) {
            printDebug(KEY_AUTHZ_FILTER, "Skip all filter annotation set in the service. Skip the filter");
            return true;
        }
        //Setting UUID
        int startingTime = getCurrentTime();
        printDebug(KEY_AUTHZ_FILTER, "Processing request via Authorization filter.");
        runtime:AuthenticationContext? authContext = runtime:getInvocationContext()?.authenticationContext;
        boolean result = true;
        if (authContext is runtime:AuthenticationContext) {
            string? authScheme = authContext?.scheme;
            // scope validation is done in authn filter for oauth2, hence we only need to
            //validate scopes if auth scheme is jwt.
            if (authScheme is string && authScheme == AUTH_SCHEME_JWT) {
                printDebug(KEY_AUTHZ_FILTER, "Auth scheme was resolved as : " + authScheme);
                //Start a new child span for the span.
                int | error | () balSpan = startSpan(BALLERINA_AUTHZ_FILTER);
                result = self.authzFilter.filterRequest(caller, request, context);
                //finishing span
                finishSpan(BALLERINA_AUTHZ_FILTER, balSpan);
            }
        }
        printDebug(KEY_AUTHZ_FILTER, "Returned with value: " + result.toString());
        string authHeader = runtime:getInvocationContext().attributes[AUTH_HEADER].toString();
        checkAndRemoveAuthHeaders(request, authHeader);
        setLatency(startingTime, context, SECURITY_LATENCY_AUTHZ);
        return result;

    }

    public function filterResponse(http:Response response, http:FilterContext context) returns boolean {
        if (context.attributes.hasKey(SKIP_ALL_FILTERS) && <boolean>context.attributes[SKIP_ALL_FILTERS]) {
            printDebug(KEY_AUTHZ_FILTER, "Skip all filter annotation set in the service. Skip the filter");
            return true;
        }
        int startingTime = getCurrentTime();
        boolean result = doAuthzFilterResponse(response, context);
        setLatency(startingTime, context, SECURITY_LATENCY_AUTHZ_RESPONSE);
        return result;
    }

};



public function doAuthzFilterResponse(http:Response response, http:FilterContext context) returns boolean {
    // In authorization filter we have specifically set the error payload since we are using ballerina in built
    // authzFilter
    // when unauthorized response is coming from a backend/interceptor,
    // we should avoid sending mgw unauthorized error response to the client
    map<any> attributes = runtime:getInvocationContext().attributes;
    // check if the response is coming from the backend.
    boolean didEpRespond = attributes.hasKey(DID_EP_RESPOND) && <boolean>attributes[DID_EP_RESPOND];
    // check if the response is coming from an interceptor.
    boolean isRespondDone = attributes.hasKey(RESPOND_DONE) && <boolean>attributes[RESPOND_DONE];

    if (response.statusCode == FORBIDDEN && !didEpRespond && !isRespondDone) {
        if (runtime:getInvocationContext().attributes[ERROR_CODE] is ()) {
            if (context.attributes[ERROR_CODE] is ()) {
                setAuthorizationFailureMessage(response, context);
            }
        }

    }
    return true;
}


public function setAuthorizationFailureMessage(http:Response response, http:FilterContext context) {
    string errorDescription = INVALID_SCOPE_MESSAGE;
    string errorMessage = INVALID_SCOPE_MESSAGE;
    int errorCode = INVALID_SCOPE;
    response.setContentType(APPLICATION_JSON);
    if (!needGrpcResponseFiltering(response, context)) {
        json payload = {
            fault: {
                code: errorCode,
                message: errorMessage,
                description: errorDescription
            }
        };
    response.setJsonPayload(payload);
    } else {
        attachGrpcErrorHeaders (response, errorMessage);
    }
}
