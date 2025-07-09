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
// Character creation types
export var CharacterClass;
(function (CharacterClass) {
    CharacterClass["Fighter"] = "fighter";
    CharacterClass["Mage"] = "mage";
    CharacterClass["Cleric"] = "cleric";
    CharacterClass["Thief"] = "thief";
    CharacterClass["Ranger"] = "ranger";
    CharacterClass["Paladin"] = "paladin";
})(CharacterClass || (CharacterClass = {}));
export var AttributeMethod;
(function (AttributeMethod) {
    AttributeMethod["Roll"] = "roll";
    AttributeMethod["PointBuy"] = "pointbuy";
    AttributeMethod["Standard"] = "standard";
    AttributeMethod["Custom"] = "custom";
})(AttributeMethod || (AttributeMethod = {}));
// Effect system types
export var EffectType;
(function (EffectType) {
    EffectType["DamageOverTime"] = "damage_over_time";
    EffectType["HealOverTime"] = "heal_over_time";
    EffectType["Poison"] = "poison";
    EffectType["Burning"] = "burning";
    EffectType["Bleeding"] = "bleeding";
    EffectType["Stun"] = "stun";
    EffectType["Root"] = "root";
    EffectType["StatBoost"] = "stat_boost";
    EffectType["StatPenalty"] = "stat_penalty";
})(EffectType || (EffectType = {}));
// Quest system types
export var QuestStatus;
(function (QuestStatus) {
    QuestStatus[QuestStatus["NotStarted"] = 0] = "NotStarted";
    QuestStatus[QuestStatus["Active"] = 1] = "Active";
    QuestStatus[QuestStatus["Completed"] = 2] = "Completed";
    QuestStatus[QuestStatus["Failed"] = 3] = "Failed";
})(QuestStatus || (QuestStatus = {}));
//# sourceMappingURL=RPCTypes.js.map