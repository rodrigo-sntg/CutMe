#!/bin/bash

# Defina a cobertura mínima desejada (em %)
MIN_COVERAGE=80.0

# Execute os testes e gere o arquivo de cobertura
echo -e "\n\033[1;34mExecutando testes e gerando cobertura...\033[0m"
go test ./... -coverprofile=coverage.out -coverpkg=./internal/application/usecase,./internal/infrastructure/aws/s3,./internal/infrastructure/aws/db,./internal/infrastructure/aws/signed_url,./internal/infrastructure/aws/sqs

# Verifique se o arquivo coverage.out foi gerado com sucesso
if [ ! -f coverage.out ]; then
  echo -e "\033[1;31mErro: Arquivo coverage.out não encontrado. Verifique se os testes foram executados corretamente.\033[0m"
  exit 1
fi

# Extraia a cobertura total
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

# Exiba o relatório de cobertura
echo -e "\n\033[1;34mRelatório de cobertura:\033[0m"
go tool cover -func=coverage.out

# Compare a cobertura total com a mínima usando awk
echo -e "\n\033[1;34mVerificando cobertura mínima...\033[0m"
if (( $(awk 'BEGIN {print ('"$TOTAL_COVERAGE"' < '"$MIN_COVERAGE"')}') )); then
  echo -e "\033[1;31mCobertura insuficiente: $TOTAL_COVERAGE%. Mínimo esperado: $MIN_COVERAGE%.\033[0m"
  exit 1
else
  echo -e "\033[1;32mCobertura suficiente: $TOTAL_COVERAGE% (Mínimo esperado: $MIN_COVERAGE%).\033[0m"
fi
