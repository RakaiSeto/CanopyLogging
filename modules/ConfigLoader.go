package modules

func LoadConfig() map[string]string {
	mapConfig := make(map[string]string)


	// Local
	mapConfig["mongoDBHost"] = "172.31.4.168"
	mapConfig["mongoDBPort"] = "27017"
	mapConfig["logEndpoint"] = "http://localhost:55555/log"

	mapConfig["rabbitHost"] = "172.31.3.176"
	mapConfig["rabbitPort"] = "5672"
	mapConfig["rabbitUser"] = "monpark"
	mapConfig["rabbitPass"] = "CakepBanget123!"
	mapConfig["rabbitVHost"] = "MONETA"

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

