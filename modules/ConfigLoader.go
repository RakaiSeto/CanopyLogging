package modules

func LoadConfig() map[string]string {
	mapConfig := make(map[string]string)


	// Local
	mapConfig["mongoDBHost"] = "localhost"
	mapConfig["mongoDBPort"] = "27017"
	mapConfig["logEndpoint"] = "http://localhost:55555/log"

	return mapConfig
}

func LoadConfigProduction() map[string]string {
	mapConfig := make(map[string]string)

	// Production
	mapConfig["mongoDBHost"] = "172.31.2.44"
	mapConfig["mongoDBPort"] = "27017"
	mapConfig["logEndpoint"] = "http://localhost:55555/log"

	return mapConfig
}

