package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault-client-go"
)

func LoadConfig() map[string]string {
	ctx := context.Background()
	mapConfig := make(map[string]string)

	cl := initVaultClient()

	err := cl.SetToken("hvs.CAESINmll07fOjkBAxz-rgCxtA6dV5OKZ9UJeQn_UDbVCHuCGh4KHGh2cy5zQ3BEc2QwZEhPeDN0dzZEaEFMNzRKa3o")
	if err != nil {
		fmt.Println("FAILED TO SET VAULT TOKEN")
		panic(err)
	}
	fmt.Println("SUCCESS TO SET VAULT TOKEN")
	
	resp, err := cl.Read(ctx, "v1/topsecret/data/canopyLogUser")
	if err != nil {
		fmt.Println("FAILED TO READ VAULT DATA")
		panic(err)
	}
	fmt.Println("SUCCESS TO READ VAULT DATA")

	data, ok := resp.Data["data"].(map[string]interface{})
	if !ok {
		panic("not map interface")
	}

	for key, value := range data {
        strKey := fmt.Sprintf("%v", key)
        strValue := fmt.Sprintf("%v", value)

        mapConfig[strKey] = strValue
    }

	return mapConfig
}

func initVaultClient() *vault.Client {
	// prepare a client with the given base address
	client, err := vault.New(
		vault.WithAddress("http://195.85.19.218:8200"),
		vault.WithRequestTimeout(10*time.Second),
	)

	if err != nil {
		fmt.Println("FAILED TO INITIATE VAULT CLIENT")
		panic(err)
	}
	fmt.Println("SUCCESS TO INITIATE VAULT CLIENT")
	
	return client
}
