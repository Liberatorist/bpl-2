// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Liberatorist",
            "email": "Liberatorist@gmail.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/events": {
            "get": {
                "description": "Fetches all events",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/controller.EventResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "description": "Creates an events",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "parameters": [
                    {
                        "description": "Event to create",
                        "name": "event",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controller.EventCreate"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/controller.EventResponse"
                        }
                    }
                }
            }
        },
        "/events/{eventId}": {
            "get": {
                "description": "Gets an event by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Event ID",
                        "name": "eventId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/controller.EventResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "Deletes an event",
                "tags": [
                    "event"
                ],
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Event ID",
                        "name": "eventId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            },
            "patch": {
                "description": "Updates an event",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "event"
                ],
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Event ID",
                        "name": "eventId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Event to update",
                        "name": "event",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controller.EventUpdate"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controller.EventResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "controller.EventCreate": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "is_current": {
                    "type": "boolean"
                },
                "max_size": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "controller.EventResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "is_current": {
                    "type": "boolean"
                },
                "max_size": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "scoring_category_id": {
                    "type": "integer"
                }
            }
        },
        "controller.EventUpdate": {
            "type": "object",
            "properties": {
                "is_current": {
                    "type": "boolean"
                },
                "max_size": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    },
    "externalDocs": {
        "description": "OpenAPI",
        "url": "https://swagger.io/resources/open-api/"
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "2.0",
	Host:             "localhost:8000",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "BPL Backend API",
	Description:      "This is the backend API for the BPL project.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
