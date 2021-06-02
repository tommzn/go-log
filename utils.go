package log

/**
func NewLogger(logLevel LogLevel, namespace string) Logger {
	return &DefaultLogger{
		logLevel:  logLevel,
		namespace: namespace,
		context:   make(map[string]string),
		logClient: log.New(os.Stdout, "", 0),
	}
}

func NewLoggerFromConfig(conf config.Config, secretsMenager secrets.SecretsManager) Logger {

	loggerType := conf.Get("log.type", config.AsStringPtr("logzio"))
	if *loggerType == "logzio" {
		return NewLogzIoLogger(conf, secretsMenager)
	} else {
		logLevelName := conf.Get("log.loglevel", config.AsStringPtr(DEFAULT_LOGLEVEL))
		logLevel := LogLevelByName(*logLevelName)
		namespace := conf.Get("log.namespace", config.AsStringPtr(DEFAULT_NAMESPACE))
		return &DefaultLogger{
			logLevel:  logLevel,
			namespace: *namespace,
			context:   make(map[string]string),
			logClient: log.New(os.Stdout, "", 0),
		}
	}
}

func NewLoggerByLogLevelName(logLevelName, namespace string) Logger {
	return NewLogger(LogLevelByName(logLevelName), namespace)
}
*/

/**
// appendLineBreakIfNecessary adds a line break to given log message
// if it does not end with one at the moment.
func appendLineBreakIfNecessary(logMessage string) string {
	if !strings.HasSuffix(logMessage, "\n") {
		logMessage += "\n"
	}
	return logMessage
}
*/

/**
// logTimeStamp returns a log timestamp in format: "2006-01-02 15:04:05.000000"
func logTimeStamp() string {
	return time.Now().Format("2006-01-02 15:04:05.000000")
}
*/
