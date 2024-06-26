# Desafio Go

## Client Server API

Neste desafio vamos aplicar o que aprendemos sobre webserver http, contextos, banco de dados e manipulação de arquivos com Go.

Você precisará nos entregar dois sistemas em Go:

- client.go
- server.go

Os requisitos para cumprir este desafio são:

- [x] O client.go deverá realizar uma requisição HTTP no server.go solicitando a cotação do dólar.
- [ ] O server.go deverá
  - [x] consumir a API contendo o câmbio de Dólar e Real no endereço: <https://economia.awesomeapi.com.br/json/last/USD-BRL>
  - [x] e em seguida deverá retornar no formato JSON o resultado para o cliente.
- [ ] Usando o package "context", o server.go deverá
  - [x] registrar no banco de dados SQLite cada cotação recebida,
  - [x] sendo que o timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms
  - [x] e o timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms.
- [x] O client.go precisará receber do server.go apenas o valor atual do câmbio (campo "bid" do JSON).
- [x] Utilizando o package "context", o client.go terá um timeout máximo de 300ms para receber o resultado do server.go.
- [ ] Os 3 contextos deverão retornar erro nos logs caso o tempo de execução seja insuficiente.
  - [ ] contexto requisição API de cotação do dólar
  - [ ] contexto persistência banco de dados
  - [ ] contexto recebimento de dados do server.go
- [ ] O client.go terá que salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
- [x] O endpoint necessário gerado pelo server.go para este desafio será: /cotacao e a porta a ser utilizada pelo servidor HTTP será a 8080.
