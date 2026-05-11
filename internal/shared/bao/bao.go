package bao

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
)

func NewBaoClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = os.Getenv("BAO_ADDR")
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Не удалось создать клиент опенбао: %v\n", err)
	}
	client.SetToken("myroot")
	for i := 0; i < 5; i++ {
		// Проверяем статус здоровья OpenBao
		health, err := client.Sys().Health()
		if err == nil && !health.Sealed {
			return client, nil
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("OpenBao не ответил после 5 попыток: %s", err)
}
