/**
 * @fileoverview gRPC-Web generated client stub for main
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');

const proto = {};
proto.main = require('./token_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.main.ServerClient =
  function (hostname, credentials, options) {
    if (!options) options = {};
    options['format'] = 'text';

    /**
     * @private @const {!grpc.web.GrpcWebClientBase} The client
     */
    this.client_ = new grpc.web.GrpcWebClientBase(options);

    /**
     * @private @const {string} The hostname
     */
    this.hostname_ = hostname;

  };


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.main.ServerPromiseClient =
  function (hostname, credentials, options) {
    if (!options) options = {};
    options['format'] = 'text';

    /**
     * @private @const {!grpc.web.GrpcWebClientBase} The client
     */
    this.client_ = new grpc.web.GrpcWebClientBase(options);

    /**
     * @private @const {string} The hostname
     */
    this.hostname_ = hostname;

  };


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.main.Token,
 *   !proto.main.Token>}
 */
const methodDescriptor_Server_AuthUser = new grpc.web.MethodDescriptor(
  '/main.Server/AuthUser',
  grpc.web.MethodType.UNARY,
  proto.main.Token,
  proto.main.Token,
  /**
   * @param {!proto.main.Token} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.Token.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.main.Token,
 *   !proto.main.Token>}
 */
const methodInfo_Server_AuthUser = new grpc.web.AbstractClientBase.MethodInfo(
  proto.main.Token,
  /**
   * @param {!proto.main.Token} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.Token.deserializeBinary
);


/**
 * @param {!proto.main.Token} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.main.Token)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.main.Token>|undefined}
 *     The XHR Node Readable Stream
 */
proto.main.ServerClient.prototype.authUser =
  function (request, metadata, callback) {
    return this.client_.rpcCall(this.hostname_ +
      '/main.Server/AuthUser',
      request,
      metadata || {},
      methodDescriptor_Server_AuthUser,
      callback);
  };


/**
 * @param {!proto.main.Token} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.main.Token>}
 *     Promise that resolves to the response
 */
proto.main.ServerPromiseClient.prototype.authUser =
  function (request, metadata) {
    return this.client_.unaryCall(this.hostname_ +
      '/main.Server/AuthUser',
      request,
      metadata || {},
      methodDescriptor_Server_AuthUser);
  };


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.main.Token,
 *   !proto.main.Pantry>}
 */
const methodDescriptor_Server_GetPantry = new grpc.web.MethodDescriptor(
  '/main.Server/GetPantry',
  grpc.web.MethodType.UNARY,
  proto.main.Token,
  proto.main.Pantry,
  /**
   * @param {!proto.main.Token} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.Pantry.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.main.Token,
 *   !proto.main.Pantry>}
 */
const methodInfo_Server_GetPantry = new grpc.web.AbstractClientBase.MethodInfo(
  proto.main.Pantry,
  /**
   * @param {!proto.main.Token} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.Pantry.deserializeBinary
);


/**
 * @param {!proto.main.Token} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.main.Pantry)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.main.Pantry>|undefined}
 *     The XHR Node Readable Stream
 */
proto.main.ServerClient.prototype.getPantry =
  function (request, metadata, callback) {
    return this.client_.rpcCall(this.hostname_ +
      '/main.Server/GetPantry',
      request,
      metadata || {},
      methodDescriptor_Server_GetPantry,
      callback);
  };


/**
 * @param {!proto.main.Token} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.main.Pantry>}
 *     Promise that resolves to the response
 */
proto.main.ServerPromiseClient.prototype.getPantry =
  function (request, metadata) {
    return this.client_.unaryCall(this.hostname_ +
      '/main.Server/GetPantry',
      request,
      metadata || {},
      methodDescriptor_Server_GetPantry);
  };


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.main.Token,
 *   !proto.main.UserInfo>}
 */
const methodDescriptor_Server_GetUserInfo = new grpc.web.MethodDescriptor(
  '/main.Server/GetUserInfo',
  grpc.web.MethodType.UNARY,
  proto.main.Token,
  proto.main.UserInfo,
  /**
   * @param {!proto.main.Token} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.UserInfo.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.main.Token,
 *   !proto.main.UserInfo>}
 */
const methodInfo_Server_GetUserInfo = new grpc.web.AbstractClientBase.MethodInfo(
  proto.main.UserInfo,
  /**
   * @param {!proto.main.Token} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.UserInfo.deserializeBinary
);


/**
 * @param {!proto.main.Token} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.main.UserInfo)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.main.UserInfo>|undefined}
 *     The XHR Node Readable Stream
 */
proto.main.ServerClient.prototype.getUserInfo =
  function (request, metadata, callback) {
    return this.client_.rpcCall(this.hostname_ +
      '/main.Server/GetUserInfo',
      request,
      metadata || {},
      methodDescriptor_Server_GetUserInfo,
      callback);
  };


/**
 * @param {!proto.main.Token} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.main.UserInfo>}
 *     Promise that resolves to the response
 */
proto.main.ServerPromiseClient.prototype.getUserInfo =
  function (request, metadata) {
    return this.client_.unaryCall(this.hostname_ +
      '/main.Server/GetUserInfo',
      request,
      metadata || {},
      methodDescriptor_Server_GetUserInfo);
  };


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.main.Token,
 *   !proto.main.Pantry>}
 */
const methodDescriptor_Server_GetGroceries = new grpc.web.MethodDescriptor(
  '/main.Server/GetGroceries',
  grpc.web.MethodType.UNARY,
  proto.main.Token,
  proto.main.Pantry,
  /**
   * @param {!proto.main.Token} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.Pantry.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.main.Token,
 *   !proto.main.Pantry>}
 */
const methodInfo_Server_GetGroceries = new grpc.web.AbstractClientBase.MethodInfo(
  proto.main.Pantry,
  /**
   * @param {!proto.main.Token} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.Pantry.deserializeBinary
);


/**
 * @param {!proto.main.Token} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.main.Pantry)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.main.Pantry>|undefined}
 *     The XHR Node Readable Stream
 */
proto.main.ServerClient.prototype.getGroceries =
  function (request, metadata, callback) {
    return this.client_.rpcCall(this.hostname_ +
      '/main.Server/GetGroceries',
      request,
      metadata || {},
      methodDescriptor_Server_GetGroceries,
      callback);
  };


/**
 * @param {!proto.main.Token} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.main.Pantry>}
 *     Promise that resolves to the response
 */
proto.main.ServerPromiseClient.prototype.getGroceries =
  function (request, metadata) {
    return this.client_.unaryCall(this.hostname_ +
      '/main.Server/GetGroceries',
      request,
      metadata || {},
      methodDescriptor_Server_GetGroceries);
  };


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.main.SearchQuery,
 *   !proto.main.Store>}
 */
const methodDescriptor_Server_GetSearchResults = new grpc.web.MethodDescriptor(
  '/main.Server/GetSearchResults',
  grpc.web.MethodType.UNARY,
  proto.main.SearchQuery,
  proto.main.Store,
  /**
   * @param {!proto.main.SearchQuery} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.Store.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.main.SearchQuery,
 *   !proto.main.Store>}
 */
const methodInfo_Server_GetSearchResults = new grpc.web.AbstractClientBase.MethodInfo(
  proto.main.Store,
  /**
   * @param {!proto.main.SearchQuery} request
   * @return {!Uint8Array}
   */
  function (request) {
    return request.serializeBinary();
  },
  proto.main.Store.deserializeBinary
);


/**
 * @param {!proto.main.SearchQuery} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.main.Store)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.main.Store>|undefined}
 *     The XHR Node Readable Stream
 */
proto.main.ServerClient.prototype.getSearchResults =
  function (request, metadata, callback) {
    return this.client_.rpcCall(this.hostname_ +
      '/main.Server/GetSearchResults',
      request,
      metadata || {},
      methodDescriptor_Server_GetSearchResults,
      callback);
  };


/**
 * @param {!proto.main.SearchQuery} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.main.Store>}
 *     Promise that resolves to the response
 */
proto.main.ServerPromiseClient.prototype.getSearchResults =
  function (request, metadata) {
    return this.client_.unaryCall(this.hostname_ +
      '/main.Server/GetSearchResults',
      request,
      metadata || {},
      methodDescriptor_Server_GetSearchResults);
  };


module.exports = proto.main;

