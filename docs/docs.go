// Code generated by swaggo/swag. DO NOT EDIT.

package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "consumes": [
        "application/json"
    ],
    "produces": [
        "application/json"
    ],
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Saad Ur Rahman",
            "url": "https://www.linkedin.com/in/saad-ur-rahman/",
            "email": "saad.ur.rahman@gmail.com"
        },
        "license": {
            "name": "GPL-3.0",
            "url": "https://opensource.org/licenses/GPL-3.0"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "This endpoint is exposed to allow load balancers etc. to check the health of the service.\nThis is achieved by the service pinging the data tier comprised of Postgres and Redis.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health healthcheck liveness"
                ],
                "summary": "Healthcheck for service liveness.",
                "operationId": "healthcheck",
                "responses": {
                    "200": {
                        "description": "message: healthy",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPSuccess"
                        }
                    },
                    "503": {
                        "description": "error message with any available details",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    }
                }
            }
        },
        "/user/delete": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Deletes a user stored in the database by marking it as deleted. The user must supply their login\ncredentials as well as complete the following confirmation message:\n\"I understand the consequences, delete my user account USERNAME HERE\"",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user users delete security"
                ],
                "summary": "Deletes a user. The user must supply their credentials as well as a confirmation message.",
                "operationId": "deleteUser",
                "parameters": [
                    {
                        "description": "The request payload for deleting an account",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.HTTPDeleteUserRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "message with a confirmation of a deleted user account",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPSuccess"
                        }
                    },
                    "400": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    },
                    "403": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    },
                    "500": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    }
                }
            }
        },
        "/user/login": {
            "post": {
                "description": "Logs in a user by validating credentials and returning a JWT.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user users login security"
                ],
                "summary": "Login a user.",
                "operationId": "loginUser",
                "parameters": [
                    {
                        "description": "Username and password to login with",
                        "name": "credentials",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UserLoginCredentials"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "a valid JWT token for the new account",
                        "schema": {
                            "$ref": "#/definitions/models.JWTAuthResponse"
                        }
                    },
                    "400": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    },
                    "409": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    },
                    "500": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    }
                }
            }
        },
        "/user/refresh": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Refreshes a user's JWT by validating it and then issuing a fresh JWT with an extended validity time.\nJWT must be expiring in under 60 seconds.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user users login refresh security"
                ],
                "summary": "Refresh a user's JWT by extending its expiration time.",
                "operationId": "loginRefresh",
                "responses": {
                    "200": {
                        "description": "A new valid JWT",
                        "schema": {
                            "$ref": "#/definitions/models.JWTAuthResponse"
                        }
                    },
                    "403": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    },
                    "500": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    },
                    "510": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    }
                }
            }
        },
        "/user/register": {
            "post": {
                "description": "Creates a user account by inserting credentials into the database. A hashed password is stored.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user users register security"
                ],
                "summary": "Register a user.",
                "operationId": "registerUser",
                "parameters": [
                    {
                        "description": "Username, password, first and last name, email address of user",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UserAccount"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "a valid JWT token for the new account",
                        "schema": {
                            "$ref": "#/definitions/models.JWTAuthResponse"
                        }
                    },
                    "400": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    },
                    "409": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    },
                    "500": {
                        "description": "error message with any available details in payload",
                        "schema": {
                            "$ref": "#/definitions/models.HTTPError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.HTTPDeleteUserRequest": {
            "type": "object",
            "required": [
                "confirmation",
                "password",
                "username"
            ],
            "properties": {
                "confirmation": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                },
                "username": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                }
            }
        },
        "models.HTTPError": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "payload": {}
            }
        },
        "models.HTTPSuccess": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "payload": {}
            }
        },
        "models.JWTAuthResponse": {
            "type": "object",
            "required": [
                "expires",
                "threshold",
                "token"
            ],
            "properties": {
                "expires": {
                    "description": "Expiration time as unix time stamp. Strictly used by client to gauge when to refresh the token.",
                    "type": "integer"
                },
                "threshold": {
                    "description": "The window in seconds before expiration during which the token can be refreshed.",
                    "type": "integer"
                },
                "token": {
                    "description": "JWT string sent to and validated by the server.",
                    "type": "string"
                }
            }
        },
        "models.UserAccount": {
            "type": "object",
            "required": [
                "email",
                "firstName",
                "lastName",
                "password",
                "username"
            ],
            "properties": {
                "email": {
                    "type": "string",
                    "maxLength": 64
                },
                "firstName": {
                    "type": "string",
                    "maxLength": 64
                },
                "lastName": {
                    "type": "string",
                    "maxLength": 64
                },
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                },
                "username": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                }
            }
        },
        "models.UserLoginCredentials": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                },
                "username": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:33723",
	BasePath:         "/api/rest/v1",
	Schemes:          []string{"http"},
	Title:            "FTeX, Inc. (Formerly Crypto-Bro's Bank, Inc.)",
	Description:      "FTeX Fiat and Cryptocurrency Banking API.\nBank, buy, and sell Fiat and Cryptocurrencies. Prices for all currencies are\nretrieved from real-time quote providers.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
