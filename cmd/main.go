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

	"github.com/lightlink/group-service/infrastructure/ws/centrifugo"
	grpcGroupDelivery "github.com/lightlink/group-service/internal/group/delivery/grpc"
	httpGroupDelivery "github.com/lightlink/group-service/internal/group/delivery/http"
	groupRepository "github.com/lightlink/group-service/internal/group/repository/postgres"
	groupUsecase "github.com/lightlink/group-service/internal/group/usecase"
	httpMessageDelivery "github.com/lightlink/group-service/internal/message/delivery/http"
	kafkaMessageFilterDelivery "github.com/lightlink/group-service/internal/message/delivery/kafka"
	messageHateSpeechRepository "github.com/lightlink/group-service/internal/message/repository/kafka"
	messageRepository "github.com/lightlink/group-service/internal/message/repository/postgres"
	messageUsecase "github.com/lightlink/group-service/internal/message/usecase"
	notificationRepository "github.com/lightlink/group-service/internal/notification/repository/kafka"
	proto "github.com/lightlink/group-service/protogen/group"
	"google.golang.org/grpc"
)

func main() {
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Ошибка при подключении к БД: %v", err)
	}
	defer db.Close()

	centrifugoKey := os.Getenv("CENTRIFUGO_HTTP_API_KEY")
	centrifugoClient := centrifugo.NewCentrifugoClient("http://centrifugo:8000", centrifugoKey)

	// === Repositories ===
	grpRepo := groupRepository.NewGroupPostgresRepository(db)
	msgRepo := messageRepository.NewMessagePostgresRepository(db)
	msgHateRepo, err := messageHateSpeechRepository.NewMessageHateSpeechRepository("kafka:29092", "input_hate_speech")
	if err != nil {
		log.Fatalf("Ошибка инициализации message hate speech repo: %v", err)
	}
	notifyRepo, err := notificationRepository.NewNotificationKafkaRepository("kafka:29092", "notifications", "http://schema_registry:9091")
	if err != nil {
		log.Fatalf("Ошибка инициализации notification repo: %v", err)
	}

	// === Usecases ===
	grpUC := groupUsecase.NewGroupUsecase(grpRepo, notifyRepo)
	msgUC := messageUsecase.NewMessageUsecase(msgRepo, notifyRepo, grpRepo, msgHateRepo, centrifugoClient)

	// === Запуск gRPC сервера ===
	go startGRPC(grpUC)

	// === Запуск HTTP сервера ===
	startHTTP(grpUC, msgUC)
}

func startGRPC(groupUsecase groupUsecase.GroupUsecaseI) {
	listener, err := net.Listen("tcp", ":8084")
	if err != nil {
		log.Fatalf("Ошибка при поднятии gRPC listener'a: %v", err)
	}

	grpcServer := grpc.NewServer()
	groupService := grpcGroupDelivery.NewGroupService(groupUsecase)
	proto.RegisterGroupServiceServer(grpcServer, groupService)

	fmt.Println("gRPC сервер запущен на порту :8084")
	log.Fatal(grpcServer.Serve(listener))
}

func startHTTP(groupUsecase groupUsecase.GroupUsecaseI, messageUsecase messageUsecase.MessageUsecaseI) {
	groupHandler := httpGroupDelivery.NewGroupHandler(groupUsecase)
	messageHandler := httpMessageDelivery.NewMessageHandler(messageUsecase)

	messageFilterConsumer, err := kafkaMessageFilterDelivery.NewMessageFilterConsumer(
		messageUsecase, "kafka:29092", "hate-speech-group", "output_hate_speech",
	)
	if err != nil {
		log.Fatalf("Ошибка запуска Kafka consumer: %v", err)
	}
	go messageFilterConsumer.Receive()

	router := mux.NewRouter()
	router.HandleFunc("/api/group/{groupID}/info", groupHandler.InfoHandler).Methods("GET")
	router.HandleFunc("/api/get-group-id/{friendID}", groupHandler.GetPersonalGroupID).Methods("GET")
	router.HandleFunc("/api/group/{groupID}/start-call", groupHandler.StartCall).Methods("POST")
	router.HandleFunc("/api/messages/{groupID}", messageHandler.GetGroupMessages).Methods("GET")
	router.HandleFunc("/api/messages", messageHandler.SendMessage).Methods("POST")

	log.Println("starting server at http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func connectToDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)
	return sql.Open("postgres", dsn)
}
