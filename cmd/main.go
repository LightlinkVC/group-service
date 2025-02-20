package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	groupDelivery "github.com/lightlink/group-service/internal/group/delivery/http"
	"github.com/lightlink/group-service/internal/group/repository/postgres"
	"github.com/lightlink/group-service/internal/group/usecase"
)

func main() {
	postgresDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	postgresConnect, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		log.Fatalf("Ошибка при подключении к БД: %v", err)
	}

	defer func() {
		if err = postgresConnect.Close(); err != nil {
			panic(err)
		}
	}()

	groupRepository := postgres.NewGroupPostgresRepository(postgresConnect)
	groupUsecase := usecase.NewGroupUsecase(groupRepository)
	groupHandler := groupDelivery.NewGroupHandler(groupUsecase)

	router := mux.NewRouter()

	router.HandleFunc("/api/groups", groupHandler.CreateGroup).Methods("POST")

	log.Println("starting server at http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
