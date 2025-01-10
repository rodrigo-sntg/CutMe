package main

import (
	"CutMe/config"
	"CutMe/domain"
	"CutMe/infrastructure"
	"CutMe/routes"
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()

	router := setupRouter()

	awsSession := setupAWSSession()
	dependencies := initializeDependencies(awsSession)

	setupRoutes(router, dependencies)

	ctx, cancel := context.WithCancel(context.Background())
	go startSQSConsumer(ctx, dependencies.SQSConsumer)
	go startServer(router)

	waitForShutdown(cancel)
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Não foi possível carregar o arquivo .env. Usando variáveis padrão do ambiente.")
	}
}

func setupRouter() *gin.Engine {
	return config.SetupRouter()
}

func setupAWSSession() *session.Session {
	return config.SetupAWSSession()
}

func initializeDependencies(awsSession *session.Session) *config.Dependencies {
	return config.InitializeDependencies(awsSession)
}

func setupRoutes(router *gin.Engine, deps *config.Dependencies) {
	routes.RegisterRoutes(router, &routes.Dependencies{
		DynamoClient:       deps.DynamoClient,
		S3Client:           deps.S3Client,
		SignedURLGenerator: deps.SignedURLGenerator,
	})
}

func startSQSConsumer(ctx context.Context, consumer *infrastructure.SQSConsumer) {
	log.Println("Iniciando consumo de mensagens SQS...")
	consumer.StartConsumption(ctx, 5)
}

func startServer(router *gin.Engine) {
	if err := router.Run(); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

func waitForShutdown(cancel context.CancelFunc) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Encerrando aplicação...")
	cancel()
}

func parseEnvInt(value string, defaultValue int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

type Dependencies struct {
	S3Client           domain.S3Client
	DynamoClient       domain.DynamoClient
	EmailNotifier      domain.Notifier
	ProcessFileUseCase domain.ProcessFileUseCase
	SQSConsumer        domain.SQSConsumer
	SignedURLGenerator domain.SignedURLGenerator
}
