package mqtt

type mqttLogger struct{}

func (log mqttLogger) Println(v ...interface{}) {
	sugarSchedLogger.Infoln(v...)
}

func (log mqttLogger) Printf(format string, v ...interface{}) {
	sugarSchedLogger.Infof(format, v...)
}
