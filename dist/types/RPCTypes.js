/**
 * JSON-RPC 2.0 type definitions for GoldBox RPG Engine
 * Provides strict typing for RPC communication with the Go backend
 */
// Standard JSON-RPC error codes
export var RPCErrorCode;
(function (RPCErrorCode) {
    RPCErrorCode[RPCErrorCode["PARSE_ERROR"] = -32700] = "PARSE_ERROR";
    RPCErrorCode[RPCErrorCode["INVALID_REQUEST"] = -32600] = "INVALID_REQUEST";
    RPCErrorCode[RPCErrorCode["METHOD_NOT_FOUND"] = -32601] = "METHOD_NOT_FOUND";
    RPCErrorCode[RPCErrorCode["INVALID_PARAMS"] = -32602] = "INVALID_PARAMS";
    RPCErrorCode[RPCErrorCode["INTERNAL_ERROR"] = -32603] = "INTERNAL_ERROR";
    // Server error range: -32099 to -32000
    RPCErrorCode[RPCErrorCode["SERVER_ERROR_MIN"] = -32099] = "SERVER_ERROR_MIN";
    RPCErrorCode[RPCErrorCode["SERVER_ERROR_MAX"] = -32000] = "SERVER_ERROR_MAX";
})(RPCErrorCode || (RPCErrorCode = {}));
//# sourceMappingURL=RPCTypes.js.map