# CutMe API

A Go-based API service for managing file uploads with AWS infrastructure integration.

## 🚀 Features

- File upload management with AWS S3
- DynamoDB integration for data storage
- Authentication middleware
- CORS configuration for frontend integration
- Docker support
- Terraform infrastructure as code

## 🛠️ Technologies

- Go 1.23+
- Gin Web Framework
- AWS (S3, DynamoDB)
- Docker
- Terraform

## 📋 Prerequisites

- Go 1.23 or higher
- Docker
- AWS CLI configured
- Terraform (for infrastructure deployment)

## 🔧 Configuration

### Local Development

1. Clone the repository:
```bash
git clone [repository-url]
cd CutMe
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables (create a `.env` file):
```env
AWS_REGION=your-region
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
```

## 🏃‍♂️ Running the Application

### Local Development
```bash
go run cmd/main.go
```

### Using Docker
```bash
docker build -t cutme-api .
docker run -p 8080:8080 cutme-api
```

## 🏗️ Infrastructure Deployment

The project uses Terraform for infrastructure management:

1. Initialize Terraform:
```bash
terraform init
```

2. Review changes:
```bash
terraform plan
```

3. Apply infrastructure:
```bash
terraform apply
```

## 🌐 API Endpoints

### Public Endpoints
- `GET /` - Welcome message

### Protected Endpoints (Requires Authentication)
- `GET /api/uploads` - List uploads
- `POST /api/upload` - Create new upload
- `POST /api/uploads/signed-url` - Generate signed URL for upload

## ⚙️ CORS Configuration

The API is configured to accept requests from:
- Origin: `http://localhost:4200`
- Methods: GET, POST, PUT, DELETE, OPTIONS
- Headers: Origin, Content-Type, Authorization

## 🔒 Security

- All API endpoints under `/api` are protected with authentication middleware
- S3 bucket is configured with private access
- CloudFront distribution is used for secure file serving

## 📝 TODO

- [ ] Create links to files
- [ ] Require authentication for file downloads

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
