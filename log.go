package selfupdate

// LogError will be called to log any reason that have prevented an executable update
var LogError func(string, ...any)

// LogInfo will be called to log any reason that prevented an executable update due to a "user" decision via one of the callback
var LogInfo func(string, ...any)

// LogDebug will be called to log any reason that prevented an executable update, because there wasn't any available detected
var LogDebug func(string, ...any)

func logError(format string, p ...any) {
	if LogError == nil {
		return
	}
	LogError(format, p...)
}

func logInfo(format string, p ...any) {
	if LogInfo == nil {
		return
	}
	LogInfo(format, p...)
}

func logDebug(format string, p ...any) {
	if LogDebug == nil {
		return
	}
	LogDebug(format, p...)
}
