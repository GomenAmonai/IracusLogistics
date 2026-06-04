package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"icaris-logistic/backend/internal/config"
	"icaris-logistic/backend/internal/db"
	"icaris-logistic/backend/internal/repository"
	"icaris-logistic/backend/internal/service"
)

// createmanager заводит менеджера из командной строки. Публичной регистрации нет
// (продукт под одну компанию), поэтому первого и последующих менеджеров создаём так:
//
//	go run ./cmd/createmanager -email=a@b.com -name="Иван" -password=secret
func main() {
	email := flag.String("email", "", "email менеджера")
	name := flag.String("name", "", "имя менеджера")
	password := flag.String("password", "", "пароль (будет захеширован bcrypt)")
	flag.Parse()

	cfg := config.Load()

	gdb, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	authService := service.NewAuthService(repository.NewManagerRepository(gdb), cfg.JWTSecret, cfg.JWTTTL)

	manager, err := authService.CreateManager(context.Background(), *email, *name, *password)
	if err != nil {
		log.Fatalf("create manager: %v", err)
	}

	fmt.Printf("создан менеджер %s <%s>, id=%s\n", manager.Name, manager.Email, manager.ID)
}
