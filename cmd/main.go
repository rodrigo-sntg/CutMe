package main

import (
	gin2 "CutMe/interface/http"
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"CutMe/config"
	"CutMe/domain/service"
)

func main() {
	loadEnv()

	router := setupRouter()
	awsSession := setupAWSSession()
	deps := initializeDependencies(awsSession)

	setupRoutes(router, deps)

	ctx, cancel := context.WithCancel(context.Background())
	go startSQSConsumer(ctx, deps.SQSConsumer)
	go startServer(router)

	waitForShutdown(cancel)
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: não foi possível carregar o arquivo .env. Usando variáveis padrão do ambiente.")
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
	// Registrando rotas via pacote "routes"
	gin2.RegisterRoutes(router, &gin2.Dependencies{
		DynamoClient:       deps.DynamoClient,
		S3Client:           deps.S3Client,
		SignedURLGenerator: deps.SignedURLGenerator,
	})
}

func startSQSConsumer(ctx context.Context, consumer service.SQSConsumer) {
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

// parseEnvInt converte string -> int, usando defaultValue em caso de erro.
func parseEnvInt(value string, defaultValue int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
