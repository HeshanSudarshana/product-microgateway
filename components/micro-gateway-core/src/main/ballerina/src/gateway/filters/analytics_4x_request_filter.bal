// Copyright (c) 2024, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
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

import ballerina/http;
import ballerina/runtime;

public type Analytics4xRequestFilter object {

    public function __init() {
        jinitAnalyticsDataPublisher();
    }

    public function filterRequest(http:Caller caller, http:Request request, http:FilterContext context) returns boolean {
        printDebug(KEY_ANALYTICS_FILTER, "Analytics Version: 4.x");
        if (context.attributes.hasKey(SKIP_ALL_FILTERS) && <boolean>context.attributes[SKIP_ALL_FILTERS]) {
            printDebug(KEY_ANALYTICS_FILTER, "Skip all filter annotation set in the service. Skipping the filter");
            return true;
        }
        //Filter only if analytics is enabled.
        if (isELKAnalyticsEnabled) {
            context.attributes[PROTOCOL_PROPERTY] = caller.protocol;
            doFilterRequest4x(request, context);
        }
        return true;
    }

    public function filterResponse(http:Response response, http:FilterContext context) returns boolean {
        printDebug(KEY_ANALYTICS_FILTER, "Analytics Version: 4.x");
        if (context.attributes.hasKey(SKIP_ALL_FILTERS) && <boolean>context.attributes[SKIP_ALL_FILTERS]) {
            printDebug(KEY_ANALYTICS_FILTER, "Skip all filter annotation set in the service. Skipping the filter");
            return true;
        }
        if (isELKAnalyticsEnabled) {
            runtime:InvocationContext invocationContext = runtime:getInvocationContext();
            boolean filterFailed = <boolean>invocationContext.attributes[FILTER_FAILED];
            printDebug(KEY_ANALYTICS_FILTER, "Filter failed filter response : " + filterFailed.toString());
            printDebug(KEY_ANALYTICS_FILTER, "Response code filter response : " + response.statusCode.toString());
            printDebug(KEY_ANALYTICS_FILTER, "Context attributes filter response : " + context.attributes.toString());
            printDebug(KEY_ANALYTICS_FILTER, "Invocation context attributes filter response : " + invocationContext.attributes.toString());
            if (context.attributes.hasKey(IS_THROTTLE_OUT)) {
                boolean isThrottleOut = <boolean>context.attributes[IS_THROTTLE_OUT];
                if (isThrottleOut) {
                  doFilterThrottleResponse4x(response, context);  
                }
                doFilterResponse4x(response, context);
            } else {
                context.attributes[THROTTLE_LATENCY] = 0;
                doFilterResponse4x(response, context);
            }
        }
        return true;
    }
};


function doFilterRequest4x(http:Request request, http:FilterContext context) {
    printDebug(KEY_ANALYTICS_FILTER, "doFilterRequest4x Mehtod called");
    error? result = trap setRequestAttributesToContext(request, context);
    if (result is error) {
        printError(KEY_ANALYTICS_FILTER, "Error while setting analytics data in request path", result);
    }
}

function doFilterFault4x(http:FilterContext context, string errorMessage) {
    printDebug(KEY_ANALYTICS_FILTER, "doFilterFault4x method called");
}

function doFilterResponse4x(http:Response response, http:FilterContext context) {
    var resp = runtime:getInvocationContext().attributes[ERROR_RESPONSE];
    printDebug(KEY_ANALYTICS_FILTER, "RESPONSE CODE HERE!!! " + response.statusCode.toString());
    printDebug(KEY_ANALYTICS_FILTER, "doFilterResponse4x method resp value : "+ resp.toString());
    Analytics4xEventData analyticsEvent = generateAnalytics4xEventData(response, context);
    jpublishAnalyticsEvent(analyticsEvent);
}

function doFilterThrottleResponse4x(http:Response response, http:FilterContext context) {
    printDebug(KEY_ANALYTICS_FILTER, "doFilterThrottleResponse4x method called");
}

function generateAnalytics4xEventData(http:Response response, http:FilterContext context) returns Analytics4xEventData {
    runtime:InvocationContext invocationContext = runtime:getInvocationContext();
    Analytics4xEventData analyticsEvent = {
        isFault: false,
        isAnonymous: false,
        isAuthenticated: true,
        responseCode: response.statusCode,
        apiUUID: "apiuuid",
        apiType: "HTTP",
        apiName: <string> invocationContext.attributes[API_NAME],
        apiVersion: <string> invocationContext.attributes[API_VERSION_PROPERTY],
        apiContext: <string> invocationContext.attributes[API_CONTEXT],
        apiCreator: <string> invocationContext.attributes[API_PUBLISHER],
        apiCreatorTenantDomain: "carbon.super",
        organizationId: "carbon.super",
        applicationUUID: "appuuid",
        applicationName: "appname",
        applicationOwner: "appowner",
        applicationKeyType: "appkeytype",
        httpMethod: "GET",
        apiResourceTemplate: "/menu",
        targetResponseCode: 200,
        responseCacheHit: false,
        destination: "destination",
        requestTime: 0,
        correlationId: "correlationid",
        regionId: "regionid",
        userAgentHeader: "useragent",
        userName: "username",
        endUserIP: "enduserip",
        backendLatency: 0,
        responseLatency: 0
    };
    return analyticsEvent;
}
