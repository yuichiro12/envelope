# envelope
.env <-> AWS Parameter Store

![envelope_list](https://user-images.githubusercontent.com/7312640/72822285-2ecaa100-3cb5-11ea-97a0-58a633438570.gif)

![envelope_apply](https://user-images.githubusercontent.com/7312640/72804122-09c53680-3c93-11ea-8941-847bb117e3a6.gif)

- all parameters are stored as `SecureString`
- `.env` is parsed by [joho/godotenv](https://github.com/joho/godotenv)

## Installation
```
go get -u github.com/yuichiro12/envelope
```

## Usage

first of all, configure aws-go-sdk with your preferable way:

https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html


#### list
list all parameters in AWS Parameter Store with given path
```
envelope list /Myservice/MyApp/Dev
```

create local `.env` from Parameter Store:
```
envelope list /Myservice/MyApp/Dev > .env
```

#### apply
apply .env to AWS Parameter Store with given prefix and filepath
```
envelope apply -f /path/to/.env /Myservice/MyApp/Dev
```

#### diff
show diff before applying .env
```
envelope diff -f /path/to/.env /Myservice/MyApp/Dev
```
