// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
// Copyright 2016-2018, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
'use strict';
var grpc = require('grpc');
var engine_pb = require('./engine_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');

function serialize_google_protobuf_Empty(arg) {
  if (!(arg instanceof google_protobuf_empty_pb.Empty)) {
    throw new Error('Expected argument of type google.protobuf.Empty');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_google_protobuf_Empty(buffer_arg) {
  return google_protobuf_empty_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pulumirpc_LogRequest(arg) {
  if (!(arg instanceof engine_pb.LogRequest)) {
    throw new Error('Expected argument of type pulumirpc.LogRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_pulumirpc_LogRequest(buffer_arg) {
  return engine_pb.LogRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// Engine is an interface into the core engine responsible for orchestrating resource operations.
var EngineService = exports.EngineService = {
  // Log logs a global message in the engine, including errors and warnings.
  log: {
    path: '/pulumirpc.Engine/Log',
    requestStream: false,
    responseStream: false,
    requestType: engine_pb.LogRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pulumirpc_LogRequest,
    requestDeserialize: deserialize_pulumirpc_LogRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
};

exports.EngineClient = grpc.makeGenericClientConstructor(EngineService);
