# CutMe API

Uma API para gerenciamento e processamento de arquivos (principalmente **vídeos**), utilizando serviços da **AWS** como S3, DynamoDB, SQS, Cognito e CloudFront.

---

## 🚀 **Features**

- Upload seguro de arquivos via AWS S3.
- Armazenamento e consulta de metadados no DynamoDB.
- **Middleware** de autenticação integrado com Cognito.
- **CORS** configurado para integração com frontend.
- **Docker** para containerização.
- Infraestrutura como código com **Terraform**.
- Downloads autenticados via **CloudFront** com Lambda@Edge.

---

## 🛠️ **Tecnologias**

- **Go** 1.23+
- **Gin** (Web Framework).
- **AWS**: S3, DynamoDB, Cognito, CloudFront, Lambda@Edge, SQS.
- **Docker** para ambientes isolados.
- **Terraform** para provisionamento de infraestrutura.

---

## 📋 **Requisitos**

- Go 1.23 ou superior.
- Docker instalado.
- AWS CLI configurado.
- Terraform para deploy da infraestrutura.

---

## 🔧 **Configuração**

### **Ambiente Local**

1. Clone o repositório:
   ```bash
   git clone https://github.com/rodrigo-sntg/CutMe.git
   cd CutMe
   ```

2. Instale as dependências:
   ```bash
   go mod download
   ```

3. Configure variáveis de ambiente criando um arquivo `.env`:
   ```env
   AWS_REGION=your-region
   REMOVIDO
   REMOVIDO
   USER_POOL_ID=your-cognito-pool-id
   ```
---

## 🏃‍♂️ **Executando a Aplicação**

### **Localmente**
```bash
go run cmd/main.go
```

### **Docker**
```bash
docker build -t cutme-api .
docker run -p 8080:8080 cutme-api
```

---

## 🏗️ **Deploy de Infraestrutura**

### Usando **Terraform**:

1. Inicialize o Terraform:
   ```bash
   terraform init
   ```

2. Veja as mudanças planejadas:
   ```bash
   terraform plan
   ```

3. Aplique as mudanças:
   ```bash
   terraform apply
   ```

---

## 🌐 **API Endpoints**

### **Públicos**
- `GET /` - Mensagem de boas-vindas.

### **Protegidos (JWT via Cognito obrigatório)**
- `GET /api/uploads` - Lista uploads do usuário.
- `POST /api/upload` - Cria novo registro de upload.
- `POST /api/uploads/signed-url` - Gera URL assinada para upload seguro no S3.

---

## ✨ **Decisões Técnicas e Arquitetura**

### **1. Linguagem: Go**

- **Alto Desempenho para Vídeos**
    - Go foi escolhido por sua eficiência em tarefas intensivas de CPU e I/O (processamento de vídeos, como extração de frames via `ffmpeg-go`).
- **Concorrência Simples**
    - Goroutines para execução paralela de múltiplos uploads.
- **Binário Único**
    - Simplicidade no deploy (comparado a runtimes como Node.js ou Java).

---

### **2. Banco de Dados: DynamoDB**

- **Escalabilidade**: Gerencia grandes volumes sem particionamento manual.
- **Modelo Simples**: Cada registro de upload inclui `id`, `status`, `URL` e timestamps.
- **Custos Otimizados**: Cobrança por RCU/WCU ou On-Demand.

---

### **3. Segurança com Cognito, CloudFront e Lambda@Edge**

#### **Autenticação com Cognito**
- Gerenciamento de identidade de usuários.
- Geração de **tokens JWT** para autenticação em:
    - **Endpoints REST** na API (validado pelo middleware Go).
    - **Download de arquivos** via CloudFront.

#### **CloudFront com Lambda@Edge**
- CloudFront serve arquivos armazenados no S3, mas com uma camada de autenticação adicional:
    - **Lambda@Edge** valida o token JWT do Cognito.
    - Tokens são verificados contra as chaves públicas baixadas de JWKS (`https://cognito-idp.<region>.amazonaws.com/<user_pool_id>/.well-known/jwks.json`).
    - Acesso negado (`401 Unauthorized`) se o token for inválido.

#### **Por que usar CloudFront?**
1. **Segurança**:
    - Sem acesso direto ao S3.
    - Autenticação obrigatória para downloads.
2. **Performance**:
    - Menor latência com entrega via **edge locations**.
3. **Fluxo Unificado**:
    - Acesso controlado, independentemente do bucket.

---

### **4. Arquitetura de Processamento de Vídeo**

1. **Upload**:
    - Cliente faz upload via URL assinada gerada pela API.
    - Arquivo é salvo em um bucket S3 privado.

2. **Processamento**:
    - **S3** envia eventos para **SQS**.
    - Workers em Go consomem as mensagens e executam:
        - **Download** do vídeo do S3.
        - Processamento via **ffmpeg-go** (ex.: extração de frames).
        - Compressão dos frames para um arquivo ZIP.
        - **Upload** do ZIP de volta ao S3.

3. **Atualização e Notificações**:
    - DynamoDB atualiza o status do upload (`PROCESSING`, `PROCESSED`, `FAILED`).
    - E-mails são enviados ao usuário ao término.

---

### **5. Detalhes do CloudFront e Lambda@Edge**

#### **Configuração do CloudFront**
- Servir arquivos via CDN com _edge locations_.
- Autenticação em cada requisição usando Lambda@Edge.

#### **Função Lambda@Edge**:
- Escrita em Node.js com `jsonwebtoken` e `jwk-to-pem`.
- Fluxo:
    1. Recebe o cabeçalho `REMOVIDO`.
    2. Busca JWKS no Cognito.
    3. Decodifica e valida o JWT.
    4. Responde com o arquivo se válido, ou `401 Unauthorized` se inválido.

---

### **6. Performance e Custos**

- **Go** vs. **Java/Node** para vídeo:
    - Melhor eficiência de memória com binário enxuto.
    - Paralelismo via goroutines reduz latência.
- **Custos AWS**:
    - **S3**: Cobrança por armazenamento e requests.
    - **CloudFront**: GB transferidos e requests.
    - **Lambda@Edge**: Requests e tempo de execução.
    - **DynamoDB**: Uso RCU/WCU.
    - **SQS**: Cobrança por polling e mensagens. (Por isso usamos long polling de 20s)

---

### **7. Segurança Avançada**

- **Criptografia**:
    - S3 com SSE-KMS.
    - DynamoDB com encriptação em repouso.
- **CORS**:
    - Regras restritas para origens confiáveis.
- **Políticas de Acesso**:
    - S3 acessível apenas via CloudFront.

---

### **8. Testes e Cobertura**

#### **Execução de Testes**
Para garantir a qualidade e cobertura dos testes, siga os passos abaixo:

1. **Rodar os Testes**
   ```bash
   go test ./... -coverprofile=coverage.out -coverpkg=./internal/application/usecase,./internal/infrastructure/aws/s3,./internal/infrastructure/aws/db,./internal/infrastructure/aws/signed_url,./internal/infrastructure/aws/sqs
   ```
    - **`-coverprofile`**: Gera um relatório da cobertura de código.
    - **`-coverpkg`**: Especifica pacotes críticos para análise de cobertura.

2. **Validar Cobertura**
   ```bash
   go tool cover -func=coverage.out
   ```
   Esse comando exibe a cobertura total e de cada função testada.

3. **Visualizar Cobertura no Navegador**
   ```bash
   go tool cover -html=coverage.out
   ```
   Abre um relatório interativo em HTML com destaque das linhas cobertas.

---

#### **Automação com `check_coverage.sh`**

Um script Bash foi criado para automatizar a validação da cobertura de testes. Ele verifica se a cobertura total atinge o limite mínimo (80% por padrão).

#### **Como Usar o Script**

1. Torne o script executável:
   ```bash
   chmod +x check_coverage.sh
   ```

2. Execute o script:
   ```bash
   ./check_coverage.sh
   ```

3. O script retorna:
    - **`✅ Cobertura suficiente`** se o limite for atingido.
    - **`❌ Cobertura insuficiente`** caso contrário.

---

### **Benefícios**
- **Automação**: Simplifica a validação de cobertura em pipelines CI/CD.
- **Qualidade do Código**: Garante que as áreas críticas da aplicação estejam cobertas.
- **Flexibilidade**: O limite pode ser ajustado conforme necessário.


### **Conclusão**

Esta solução combina:

- **Desempenho**: Processamento rápido com Go.
- **Segurança**: Tokens JWT, Lambda@Edge e controle granular.
- **Escalabilidade**: Infraestrutura com S3, SQS e DynamoDB.

