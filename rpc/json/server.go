// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"encoding/json"
	"net/http"

	"code.google.com/p/gorilla/rpc"
)

var null = json.RawMessage([]byte("null"))

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// serverRequest represents a JSON-RPC request received by the server.
type serverRequest struct {
	// A String containing the name of the method to be invoked.
	Method string `json:"method"`
	// An Array of objects to pass as arguments to the method.
	Params *json.RawMessage `json:"params"`
	// The request id. This can be of any type. It is used to match the
	// response with the request that it is replying to.
	Id *json.RawMessage `json:"id"`
}

// serverResponse represents a JSON-RPC response returned by the server.
type serverResponse struct {
	// The Object that was returned by the invoked method. This must be null
	// in case there was an error invoking the method.
	Result interface{} `json:"result"`
	// An Error object if there was an error invoking the method. It must be
	// null if there was no error.
	Error interface{} `json:"error"`
	// This must be the same id as the request it is responding to.
	Id *json.RawMessage `json:"id"`
}

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// NewCodec returns a new JSON Codec.
func NewCodec() *Codec {
	return &Codec{}
}

// Codec creates a CodecRequest to process each request.
type Codec struct {
}

// NewRequest returns a CodecRequest.
func (c *Codec)	NewRequest() rpc.CodecRequest {
	return new(CodecRequest)
}

// ----------------------------------------------------------------------------
// CodecRequest
// ----------------------------------------------------------------------------

// CodecRequest decodes and encodes a single request.
type CodecRequest struct {
	server  *rpc.Server
	request *serverRequest
	err     error
}

// Method returns the RPC method for the current request.
//
// The method uses a dotted notation as in "Service.Method".
func (c *CodecRequest) Method(r *http.Request) (string, error) {
	if err := c.createServerRequest(r); err != nil {
		return "", err
	}
	return c.request.Method, nil
}

// ReadRequest fills the request object for the RPC method.
func (c *CodecRequest) ReadRequest(r *http.Request, args interface{}) error {
	if err := c.createServerRequest(r); err != nil {
		return err
	}
	// JSON params is array value. RPC params is struct.
	// Unmarshal into array containing the request struct.
	params := [1]interface{}{args}
	return json.Unmarshal(*c.request.Params, &params)
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
//
// The err parameter is the error resulted from calling the RPC method,
// or nil if there was no error.
func (c *CodecRequest) WriteResponse(w http.ResponseWriter, reply interface{}, methodErr error) error {
	res := &serverResponse{
		Result: reply,
		Error:  methodErr,
		Id:     c.request.Id,
	}
	if methodErr != nil {
		// Result must be null if there was an error invoking the method.
		// http://json-rpc.org/wiki/specification#a1.2Response
		res.Result = &null
	}
	if c.request.Id == nil {
		// Id is null for notifications and they don't have a response.
		res.Id = &null
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(res)
	}
	return nil
}

// createServerRequest pre-process the request
func (c *CodecRequest) createServerRequest(r *http.Request) error {
	// Decode the request body and check if RPC method is valid.
	if c.err != nil {
		return c.err
	}
	if c.request == nil {
		defer r.Body.Close()
		req := new(serverRequest)
		if c.err = json.NewDecoder(r.Body).Decode(req); c.err != nil {
			return c.err
		}
		c.request = req
	}
	return nil
}
