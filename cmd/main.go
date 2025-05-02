package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/centrifugal/centrifuge"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	grpcGroupDelivery "github.com/lightlink/group-service/internal/group/delivery/grpc"
	httpGroupDelivery "github.com/lightlink/group-service/internal/group/delivery/http"
	groupRepository "github.com/lightlink/group-service/internal/group/repository/postgres"
	groupUsecase "github.com/lightlink/group-service/internal/group/usecase"
	httpMessageDelivery "github.com/lightlink/group-service/internal/message/delivery/http"
	kafkaMessageFilterDelivery "github.com/lightlink/group-service/internal/message/delivery/kafka"
	messageHateSpeechRepository "github.com/lightlink/group-service/internal/message/repository/kafka"
	messageRepository "github.com/lightlink/group-service/internal/message/repository/postgres"
	messageUsecase "github.com/lightlink/group-service/internal/message/usecase"
	"github.com/lightlink/group-service/internal/middleware"
	notificationRepository "github.com/lightlink/group-service/internal/notification/repository/kafka"
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

func isValidGroupChannel(channel string) bool {
	return len(channel) > 6 && channel[:6] == "group:"
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

	node, err := centrifuge.New(centrifuge.Config{})
	if err := node.Run(); err != nil {
		panic(err)
	}

	node.OnConnect(func(client *centrifuge.Client) {
		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			if !isValidGroupChannel(e.Channel) {
				fmt.Println("Invalid channel: ", e.Channel)
				cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
				return
			}

			cb(centrifuge.SubscribeReply{
				Options: centrifuge.SubscribeOptions{
					EmitPresence:  true,
					EmitJoinLeave: true,
				},
			}, nil)
		})
	})

	groupRepository := groupRepository.NewGroupPostgresRepository(postgresConnect)
	messageRepository := messageRepository.NewMessagePostgresRepository(postgresConnect)
	messageHateSpeechRepository, err := messageHateSpeechRepository.NewMessageHateSpeechRepository(
		"kafka:29092",
		"input_hate_speech",
	)
	if err != nil {
		panic(err)
	}
	notificationRepository, err := notificationRepository.NewNotificationKafkaRepository(
		"kafka:29092",
		"notifications",
		"http://schema_registry:9091",
	)
	if err != nil {
		panic(err)
	}

	groupUsecase := groupUsecase.NewGroupUsecase(groupRepository)
	messageUsecase := messageUsecase.NewMessageUsecase(
		messageRepository,
		notificationRepository,
		groupRepository,
		messageHateSpeechRepository,
	)

	groupHandler := httpGroupDelivery.NewGroupHandler(groupUsecase)
	messageHandler := httpMessageDelivery.NewMessageHandler(messageUsecase, node)

	messageFilterConsumer, err := kafkaMessageFilterDelivery.NewMessageFilterConsumer(
		messageUsecase,
		"kafka:29092",
		"hate-speech-group",
		"output_hate_speech",
	)
	if err != nil {
		panic(err)
	}

	go messageFilterConsumer.Receive()

	router := mux.NewRouter()

	/*TODO решить проблему с CORS*/
	wsHandler := centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
	router.Handle("/connection/websocket", middleware.ValidateAuthWS(wsHandler))

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
