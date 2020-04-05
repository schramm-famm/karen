# Karen
Karen is responsible for CRUD operations on user accounts. It uses MariaDB to
store the user data, including salted & hashed passwords using
[bcrypt](https://en.wikipedia.org/wiki/Bcrypt).

## Development Dependencies
- [`golang`](https://golang.org/dl/)
- [`mariadb`](https://mariadb.org/download/)
- [`docker`](https://docs.docker.com/install/)
- [`terraform`](https://www.terraform.io/downloads.html)

## Environment Variables
* `KAREN_DB_USERNAME`: username for accessing MariaDB
* `KAREN_DB_PASSWORD`: password for accessing MariaDB
* `KAREN_DB_LOCATION`: host and port where MariaDB is located (ex: "localhost:3306")

## Running Karen
1. Install the development dependencies.
2. `karen` can be run normally or in docker:  
   a. To run normally, run `make run [ENV_VAR=<VALUE>...]`  
   b. To run in docker, run `make docker-run [ENV_VAR=<VALUE>...]`

## Terraform
To deploy `karen` and its direct dependencies in AWS, use `terraform`:
1. Change directories to the `terraform/` directory.
2. Create a file in the `terraform/` directory called `terraform.tfvars` and
fill it out like this:
```
name          = "<WHATEVER_UNIQUE_NAME_YOU_WANT>"
access_key    = "<YOUR_AWS_ACCESS_KEY_ID>"
secret_key    = "<YOUR_AWS_SECRET_ACCESS_KEY>"
region        = "<AWS_REGION>" // optional
rds_username  = "<USERNAME>"
rds_password  = "<PASSWORD>"
container_tag = "<KAREN_CONTAINER_TAG_IN_ECR>" // optional
```
3. Run `terraform init` to initialize the Terraform working directory.
4. Run `terraform plan` to see what resources will be created and then
   `terraform apply` to create the resources. Enter `yes` when prompted by
   Terraform.
5. Once the previous command is done running, the AWS resources should now be
   visible in the AWS console (UI) and ready to be used for development/testing.
   Once you're done using the resources, run `terraform destroy`. Enter `yes`
   when prompted by Terraform.

## Testing
### Unit Testing
To run the Go unit tests, execute `make test`.

### Integration Testing
1. Deploy `karen` either locally or in AWS and get its endpoint.
2. Run `npm i` to install the Node.js test dependencies.
3. Run `export HOST=<YOUR_KAREN_ENDPOINT>` to tell the tests what endpoint to
   make requests against.
4. Run `npm test` to run the integration tests.

## API Documentation
All the following APIs except for `POST /karen/v1/users` and
`POST /karen/v1/users/auth` are protected by `heimdall`, so requests must have
the `Authorization` header set to the value `Bearer <token>`, where `<token>` is
the token generated by `heimdall`. `heimdall` will then forward the request with
an added `User-ID` header with the user ID value from the token. This user is
treated as the "session user" for these authenticated requests.

## APIS
### `POST /karen/v1/users`
Creates a new user.
#### Request body format
```
{
    "name": "John Smith",
    "email": "johnsmith@example.com",
    "password": "jsmithpass123"
}
```

#### Response format
`201 Created`
```
{
    "id": 1,
    "name": "John Smith",
    "email": "johnsmith@example.com",
    "avatar_url": ""
}
```

Notable error codes: `409 Conflict`

### `POST /karen/v1/users/auth`
Authenticates user credentials.
#### Request body format
```
{
    "email": "johnsmith@example.com",
    "password": "jsmithpass123"
}
```

#### Response format
`200 OK`
```
{
    "id": 1,
    "name": "John Smith",
    "email": "johnsmith@example.com"
}
```

Notable error codes: `401 Unauthorized`, `404 Not Found`

### `GET /karen/v1/users?includes=id,name,email,avatar_url`
Retrieves the session user (based on the "User-ID" header).
#### Response format
`200 OK`
```
{
    "name": "John Smith",
    "email": "johnsmith@example.com",
    "avatar_url": ""
}
```

Notable error codes: `404 Not Found`

### `GET /karen/v1/users/{user-id}?includes=id,name,email,avatar_url`
Retrieves a specified user (based on the URL path variable).
#### Response format
`200 OK`
```
{
    "name": "Jane Doe",
    "email": "janedoe@example.com",
    "avatar_url": "example.com/profile.png"
}
```

Notable error codes: `404 Not Found`

### `GET /karen/v1/users?email=<EMAIL>`
Retrieves a specified user (based on the `email` query parameter).
#### Response format
`200 OK`
```
{
    "id": 2,
    "name": "Jane Doe",
    "email": "janedoe@example.com",
    "avatar_url": "example.com/profile.png"
}
```

Notable error codes: `404 Not Found`

### `PATCH /karen/v1/users/self`
Updates the session user.
#### Request body format
```
{
    "name": "Johnny",
    "email": "newjohnsmith@example.com",
    "password": "password123"
}
```

#### Response format
`200 OK`
```
{
    "name": "Jonny",
    "email": "newjohnsmith@example.com",
    "password": "password123"
}
```

Notable error codes: `404 Not Found`, `409 Conflict`

### `DELETE /karen/v1/users/self`
Deletes the session user.
#### Response format
`204 No Content`

Notable error codes: `404 Not Found`
