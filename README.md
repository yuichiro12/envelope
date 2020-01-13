# envelope
.env <-> aws parameter store

- all parameters are stored as `SecureString`
- `.env` is parsed by [joho/godotenv](https://github.com/joho/godotenv)

## Installation
```
go get -u github.com/yuichiro12/envelope
```

## Usage

#### list
list all parameters in aws parameter store with given path
```
envelope list /Myservice/MyApp/Dev
```

create local `.env` from parameter store:
```
envelope list /Myservice/MyApp/Dev > .env
```

#### apply
apply .env to aws parameter store with given prefix and filepath
```
envelope apply -f /path/to/.env /Myservice/MyApp/Dev
```

#### diff
show diff before applying .env
```
envelope diff -f /path/to/.env /Myservice/MyApp/Dev
```
