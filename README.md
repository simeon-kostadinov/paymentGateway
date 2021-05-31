# ProcessOut Payment Gateway API Challenge

The API is created as part of an interview challenge for ProcessOut.

## Assumptions

1. The currency will be sent as for example "USD". As this solution uses an API to get the actual currencies note that not all will be supported
2. Amount in the capture will be in the account currency upon covnersion
3. When calling /authorize from the API the Bank will have a functionality to freeze the authorized amount and unfreez it after
In a real scenario, this will be an API to interact with the Bank Account of course, made it here just to make things easier and simple given the time frame of the challenge and my own availability, appologies for not making it as accurate to a real life scenario as possible.
4. Assumed that in real life there will be fees for paying in other currencies, but it's not scoped for this challenge.

## Considerations:

1. Should encrypt sensitive data when storing or using (get it encrypted and store it encrypted) (e.g. cardNumber in Auth object (needed to update bank account)) or refactor the code so it is not needed to store it.
2. Refactor some of the code that is reused in refund/void/capture funcs
3. Create a proper merchant user base system (login system with merchantIds, passwords etc. for extra security)
4. Could rework the design to be more clean and efficient
5. Could implement a way to check for duplicate authorization requests (maybe ID generation can be tied into this i.e. duplicatee authorizations will generate the same ID)
6. Can extract more string literals and values to constants if there are missed anywhere
7. Can make it so that the authorization ID is sent not in body, but via route link (similary like we send the merchant_id)
8. Started setting it up to be deployed on Kubernetes and possibly host it on AWS
9. Use gomock to test and mock the behaviour of the models interface in combination with goconvey for more efficient and robust testing
10. Extend the test beyong unit testing (maybe add EUTs, Integration tests, etc.)
11. Use Kafka to listen for incoming messages and trigger events based on it (use gRPC & protobuf as well to transmit messages)
12. Log important events and stash them in a log files. Maybe use an aggregator to display the logs (Slunk is a nice example). Makes debugging easier. Segregate log message by type e.g. info/warn/debug and by call/component (e.g. card, authorization, authentication, etc.)

## Tech Stack and Dependencies:

- Go Lang
- PostgresSQL
- Gorm ORM library
- Gorilla Mux http router
- godotenv 
- goconvey
- JWT token for authentication
- Postman for running API calls
- Docker

## Instalation and Running (local):

```bash
run: go mod download
run: go run main.go
test: go test -v 
```

## Instalation and Running (Docker):

```bash
run: docker-compose up -d 
stop: docker-compose down --remove-orphans --volumes 
test: docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

## Usage
1. Loging (POST) in with a MerchantId (system is currently just generating the JWT token and is not extended to handle proper logins and validation). MerchantId will be reused with token for authentication and API calls later.

- Endpoint:
```bash
localhost:8080/login
```

- Payload:
```json
{
    "mid" : "123456",
}
```

- Output:
```json
{
    "token" : "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJleHAiOjE2MjI0MTQ4ODgsIm1lcmNoYW50X2lkIjoxMjM0NTZ9.b-QH3zFydpVMvkNaebvHEkctRMdHCp2HcinT_tf7jfk",
}
```

2. Authorization (PUT):

- Token needs to go in Postman -> Authorization -> Type: Bearer Token

- Endpoint:
```bash
localhost:8080/{mid}/authorize
```

- Payload:
```json
{
    "cardNumber" : "4000000000000119",
    "currency": "USD",
    "cvv":"123",
    "amount": 10,
    "expirationMonth": 1,
    "expirationYear": 23
}
```

- Output:
```json
{
    "id" : "1tGYTrSLQ8JqJzy7K7cExJurbXs",
    "currency": "USD",
    "amountAvailable":10
}
```

3. Capture (POST):

- Token needs to go in Postman -> Authorization -> Type: Bearer Token

- Endpoint:
```bash
localhost:8080/{mid}/capture
```

- Payload:
```json
{
    "id" : "1tGYTrSLQ8JqJzy7K7cExJurbXs",
    "amount": 5,
    "final":false
}
```

- Output:
```json
{
    "id" : "1tGYTrSLQ8JqJzy7K7cExJurbXs",
    "currency": "USD",
    "amountAvailable":5
}
```

4. Refund (POST):

- Token needs to go in Postman -> Authorization -> Type: Bearer Token

- Endpoint:
```bash
localhost:8080/{mid}/refund
```

- Payload:
```json
{
    "id" : "1tGYTrSLQ8JqJzy7K7cExJurbXs",
    "amount": 5,
    "final":false
}
```

- Output:
```json
{
    "id" : "1tGYTrSLQ8JqJzy7K7cExJurbXs",
    "currency": "USD",
    "amountAvailable":10
}
```

5. Void (POST):

- Token needs to go in Postman -> Authorization -> Type: Bearer Token

- Endpoint:
```bash
localhost:8080/{mid}/void
```

- Payload:
```json
{
    "id" : "1tGYTrSLQ8JqJzy7K7cExJurbXs",
}
```

- Output:
```json
{
    "id" : "1tGYTrSLQ8JqJzy7K7cExJurbXs",
    "currency": "USD",
    "amountAvailable":0
}
```


