// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/orders/{order_id}/events": {
            "get": {
                "description": "Get events for a specific order",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "orders"
                ],
                "summary": "Get order events",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "Order ID",
                        "name": "order_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.OrderEvent"
                            }
                        }
                    }
                }
            }
        },
        "/webhooks/payments/orders": {
            "post": {
                "description": "handle event from JustPay!",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "webhooks"
                ],
                "summary": "handle event from JustPay!",
                "parameters": [
                    {
                        "description": "Event",
                        "name": "event",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http.postEventRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/domain.OrderEvent"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "domain.OrderEvent": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "event_id": {
                    "type": "string"
                },
                "is_final": {
                    "type": "boolean"
                },
                "order_id": {
                    "type": "string"
                },
                "order_status": {
                    "$ref": "#/definitions/domain.OrderStatus"
                },
                "updated_at": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "domain.OrderStatus": {
            "type": "string",
            "enum": [
                "cool_order_created",
                "sbu_verification_pending",
                "confirmed_by_mayor",
                "chinazes",
                "changed_my_mind",
                "failed",
                "give_my_money_back"
            ],
            "x-enum-varnames": [
                "CoolOrderCreated",
                "SbuVerificationPending",
                "ConfirmedByMayor",
                "Chinazes",
                "ChangedMyMind",
                "Failed",
                "GiveMyMoneyBack"
            ]
        },
        "http.postEventRequest": {
            "type": "object",
            "required": [
                "created_at",
                "event_id",
                "order_id",
                "order_status",
                "updated_at",
                "user_id"
            ],
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "event_id": {
                    "type": "string"
                },
                "order_id": {
                    "type": "string"
                },
                "order_status": {
                    "type": "string",
                    "enum": [
                        "cool_order_created",
                        "sbu_verification_pending",
                        "confirmed_by_mayor",
                        "changed_my_mind",
                        "failed",
                        "chinazes",
                        "give_my_money_back"
                    ]
                },
                "updated_at": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
