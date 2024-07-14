# Webmonitor Webhook Go

Este projeto é um monitor de Webhook desenvolvido em Golang.

## Instruções de Uso

1. Execute o comando abaixo no terminal para iniciar o servidor:

   ```sh
   go run main.go

2. Para monitorar um endpoint específico, acesse a URL no formato abaixo no seu navegador, onde {endpoint} é o endpoint que você deseja monitorar:

    http://localhost:8080/webhook/{endpoint}

3. Para enviar webhooks, faça requisições POST para o endpoint específico com o conteúdo desejado:

    http://localhost:8080/webhook/{endpoint}

## Exemplo de Funcionamento

Acesse a URL http://localhost:8080/webhook/12345 com uma requisição GET no navegador. Você verá uma página de monitoramento específica para o endpoint "12345".

Esta página exibirá em tempo real todas as requisições POST feitas para http://localhost:8080/webhook/12345.

## Características

Cada endpoint terá sua própria página de monitoramento independente.

As páginas de monitoramento mostram apenas as requisições relevantes para aquele endpoint específico.

Esta implementação permite monitorar múltiplos endpoints de webhook de forma independente.

Facilita o acompanhamento e o debug de diferentes fontes de webhooks.