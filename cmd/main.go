package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	grpcGroupDelivery "github.com/lightlink/group-service/internal/group/delivery/grpc"
	httpGroupDelivery "github.com/lightlink/group-service/internal/group/delivery/http"
	groupRepository "github.com/lightlink/group-service/internal/group/repository/postgres"
	groupUsecase "github.com/lightlink/group-service/internal/group/usecase"
	httpMessageDelivery "github.com/lightlink/group-service/internal/message/delivery/http"
	messageRepository "github.com/lightlink/group-service/internal/message/repository/postgres"
	messageUsecase "github.com/lightlink/group-service/internal/message/usecase"
	proto "github.com/lightlink/group-service/protogen/group"
	"google.golang.org/grpc"
)

func main() {
	// === Запускаем gRPC ===
	go startGRPC()

	// === Запускаем HTTP ===
	startHTTP()
}

func startGRPC() {
	listener, err := net.Listen("tcp", ":8084")
	if err != nil {
		log.Fatalf("Ошибка при поднятии gRPC listener'a: %v", err)
	}

	grpcServer := grpc.NewServer()

	postgresConnect, err := connectToDB()
	if err != nil {
		log.Fatalf("Ошибка при подключении к БД: %v", err)
	}
	defer func() {
		if err = postgresConnect.Close(); err != nil {
			panic(err)
		}
	}()

	groupRepository := groupRepository.NewGroupPostgresRepository(postgresConnect)
	groupUsecase := groupUsecase.NewGroupUsecase(groupRepository)
	groupService := grpcGroupDelivery.NewGroupService(groupUsecase)

	proto.RegisterGroupServiceServer(grpcServer, groupService)

	fmt.Println("gRPC сервер запущен на порту :8084")
	log.Fatal(grpcServer.Serve(listener))
}

func startHTTP() {
	postgresConnect, err := connectToDB()
	if err != nil {
		log.Fatalf("Ошибка при подключении к БД: %v", err)
	}
	defer func() {
		if err = postgresConnect.Close(); err != nil {
			panic(err)
		}
	}()

	groupRepository := groupRepository.NewGroupPostgresRepository(postgresConnect)
	messageRepository := messageRepository.NewMessagePostgresRepository(postgresConnect)

	groupUsecase := groupUsecase.NewGroupUsecase(groupRepository)
	messageUsecase := messageUsecase.NewMessageUsecase(messageRepository)

	groupHandler := httpGroupDelivery.NewGroupHandler(groupUsecase)
	messageHandler := httpMessageDelivery.NewMessageHandler(messageUsecase)

	router := mux.NewRouter()

	router.HandleFunc("/api/get-group-id/{friendID}", groupHandler.GetPersonalGroupID).Methods("GET")

	router.HandleFunc("/api/messages/{groupID}", messageHandler.GetGroupMessages).Methods("GET")
	router.HandleFunc("/api/messages", messageHandler.SendMessage).Methods("POST")

	log.Println("starting server at http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func connectToDB() (*sql.DB, error) {
	postgresDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return nil, err
	}

	return db, nil
}
