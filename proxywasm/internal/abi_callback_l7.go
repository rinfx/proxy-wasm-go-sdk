// Copyright 2020-2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

//export proxy_on_request_headers
func proxyOnRequestHeaders(contextID uint32, numHeaders int, endOfStream bool) types.Action {
	ctx, ok := currentState.httpContexts[contextID]
	if !ok {
		panic("invalid context on proxy_on_request_headers")
	}

	currentState.setActiveContextID(contextID)
	return ctx.OnHttpRequestHeaders(numHeaders, endOfStream)
}

//export proxy_on_request_body
func proxyOnRequestBody(contextID uint32, bodySize int, endOfStream bool) types.Action {
	ctx, ok := currentState.httpContexts[contextID]
	if !ok {
		panic("invalid context on proxy_on_request_body")
	}
	currentState.setActiveContextID(contextID)
	return ctx.OnHttpRequestBody(bodySize, endOfStream)
}

//export proxy_on_request_trailers
func proxyOnRequestTrailers(contextID uint32, numTrailers int) types.Action {
	ctx, ok := currentState.httpContexts[contextID]
	if !ok {
		panic("invalid context on proxy_on_request_trailers")
	}
	currentState.setActiveContextID(contextID)
	return ctx.OnHttpRequestTrailers(numTrailers)
}

//export proxy_on_response_headers
func proxyOnResponseHeaders(contextID uint32, numHeaders int, endOfStream bool) types.Action {
	ctx, ok := currentState.httpContexts[contextID]
	if !ok {
		panic("invalid context id on proxy_on_response_headers")
	}
	currentState.setActiveContextID(contextID)
	return ctx.OnHttpResponseHeaders(numHeaders, endOfStream)
}

//export proxy_on_response_body
func proxyOnResponseBody(contextID uint32, bodySize int, endOfStream bool) types.Action {
	ctx, ok := currentState.httpContexts[contextID]
	if !ok {
		panic("invalid context id on proxy_on_response_headers")
	}
	currentState.setActiveContextID(contextID)
	return ctx.OnHttpResponseBody(bodySize, endOfStream)
}

//export proxy_on_response_trailers
func proxyOnResponseTrailers(contextID uint32, numTrailers int) types.Action {
	ctx, ok := currentState.httpContexts[contextID]
	if !ok {
		panic("invalid context id on proxy_on_response_headers")
	}
	currentState.setActiveContextID(contextID)
	return ctx.OnHttpResponseTrailers(numTrailers)
}

//export proxy_on_http_call_response
func proxyOnHttpCallResponse(pluginContextID, calloutID uint32, numHeaders, bodySize, numTrailers int) {
	root, ok := currentState.pluginContexts[pluginContextID]
	if !ok {
		panic("http_call_response on invalid plugin context")
	}

	cb := root.httpCallbacks[calloutID]
	if cb == nil {
		panic("invalid callout id")
	}

	ctxID := cb.callerContextID
	currentState.setActiveContextID(ctxID)
	delete(root.httpCallbacks, calloutID)

	// Check if the context is already deleted.
	// For example, if the connection expired before the call response arrival,
	// proxy_on_http_call_response is called AFTER ProxyOnDelete is called for the context id.
	// In that case, if the callback continues response or make local reply, then the subsequent
	// callbacks (for example OnHttpResponseHeaders) would follow and result in calling callback
	// for already-deleted context id. See https://github.com/tetratelabs/proxy-wasm-go-sdk/issues/261 for detail.
	if _, ok := currentState.contextIDToRootID[ctxID]; ok {
		ProxySetEffectiveContext(ctxID)
		cb.callback(numHeaders, bodySize, numTrailers)
	}
}

//export proxy_on_redis_call_response
func proxyOnRedisCallResponse(pluginContextID, calloutID uint32, status, responseSize int) {
	root, ok := currentState.pluginContexts[pluginContextID]
	if !ok {
		panic("redis_call_response on invalid plugin context")
	}

	cb := root.redisCallbacks[calloutID]
	if cb == nil {
		panic("invalid callout id")
	}

	ctxID := cb.callerContextID
	currentState.setActiveContextID(ctxID)
	delete(root.redisCallbacks, calloutID)

	// Check if the context is already deleted.
	// For example, if the connection expired before the call response arrival,
	// proxy_on_http_call_response is called AFTER ProxyOnDelete is called for the context id.
	// In that case, if the callback continues response or make local reply, then the subsequent
	// callbacks (for example OnHttpResponseHeaders) would follow and result in calling callback
	// for already-deleted context id. See https://github.com/tetratelabs/proxy-wasm-go-sdk/issues/261 for detail.
	if _, ok := currentState.contextIDToRootID[ctxID]; ok {
		ProxySetEffectiveContext(ctxID)
		cb.callback(status, responseSize)
	}
}