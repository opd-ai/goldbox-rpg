Add structured logging to this function using logrus:

log.WithFields(logrus.Fields{
  "func": "FunctionName",
  "params": params,
}).Debug("entry")

Use these levels:
Debug - function entry/exit
Info - key operations
Warn - edge cases
Error - with err details

Format: include "func" field in every log