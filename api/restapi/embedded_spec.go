// Code generated by go-swagger; DO NOT EDIT.

package restapi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
)

var (
	// SwaggerJSON embedded version of the swagger document used at generation time
	SwaggerJSON json.RawMessage
	// FlatSwaggerJSON embedded flattened version of the swagger document used at generation time
	FlatSwaggerJSON json.RawMessage
)

func init() {
	SwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "description": "this document describes the functions, parameters and return values of the vanus operator API\n",
    "title": "Vanus Operator Interface Document",
    "version": "v0.2.0"
  },
  "basePath": "/api/v1",
  "paths": {
    "/cluster/": {
      "get": {
        "description": "get Cluster",
        "tags": [
          "cluster"
        ],
        "operationId": "getCluster",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Cluster_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "post": {
        "description": "create Cluster",
        "tags": [
          "cluster"
        ],
        "operationId": "createCluster",
        "parameters": [
          {
            "description": "需要创建的Cluster信息",
            "name": "create",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "$ref": "#/definitions/Cluster_create"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Cluster_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "delete": {
        "description": "delete Cluster",
        "tags": [
          "cluster"
        ],
        "operationId": "deleteCluster",
        "parameters": [
          {
            "description": "delete Cluster",
            "name": "delete",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "$ref": "#/definitions/Cluster_delete"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Cluster_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "patch": {
        "description": "patch Cluster",
        "tags": [
          "cluster"
        ],
        "operationId": "patchCluster",
        "parameters": [
          {
            "description": "patch info",
            "name": "patch",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "$ref": "#/definitions/Cluster_patch"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Cluster_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      }
    },
    "/connector/{name}": {
      "get": {
        "description": "get Connector",
        "tags": [
          "connector"
        ],
        "operationId": "getConnector",
        "parameters": [
          {
            "type": "string",
            "description": "connector name",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Connector_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      }
    },
    "/connectors/": {
      "get": {
        "description": "list Connector",
        "tags": [
          "connector"
        ],
        "operationId": "listConnector",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "$ref": "#/definitions/Connector_info"
                  }
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "post": {
        "description": "create Connector",
        "tags": [
          "connector"
        ],
        "operationId": "createConnector",
        "parameters": [
          {
            "description": "Connector info",
            "name": "connector",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "$ref": "#/definitions/Connector_info"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Connector_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "delete": {
        "description": "delete Connector",
        "tags": [
          "connector"
        ],
        "operationId": "deleteConnector",
        "parameters": [
          {
            "type": "string",
            "description": "connector name",
            "name": "name",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Connector_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      }
    },
    "/healthz/": {
      "get": {
        "description": "healthz",
        "tags": [
          "healthz"
        ],
        "operationId": "healthz",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Health_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      }
    }
  },
  "definitions": {
    "APIResponse": {
      "type": "object",
      "properties": {
        "code": {
          "type": "integer",
          "format": "int32"
        },
        "data": {
          "type": "object"
        },
        "message": {
          "type": "string"
        }
      }
    },
    "Cluster_create": {
      "description": "cluster create params",
      "type": "object",
      "properties": {
        "controller_replicas": {
          "description": "controller replicas",
          "type": "integer",
          "format": "int32"
        },
        "controller_storage_size": {
          "description": "controller storage size",
          "type": "string"
        },
        "store_replicas": {
          "description": "store replicas",
          "type": "integer",
          "format": "int32"
        },
        "store_storage_size": {
          "description": "store storage size",
          "type": "string"
        },
        "version": {
          "description": "cluster version",
          "type": "string"
        }
      }
    },
    "Cluster_delete": {
      "description": "cluster create params",
      "type": "object",
      "properties": {
        "force": {
          "description": "true means force delete, default false",
          "type": "boolean",
          "default": false
        }
      }
    },
    "Cluster_info": {
      "description": "cluster info",
      "type": "object",
      "properties": {
        "status": {
          "description": "cluster status",
          "type": "string"
        },
        "version": {
          "description": "cluster version",
          "type": "string"
        }
      }
    },
    "Cluster_patch": {
      "description": "cluster patch params",
      "type": "object",
      "properties": {
        "controller_replicas": {
          "description": "controller replicas",
          "type": "integer",
          "format": "int32"
        },
        "store_replicas": {
          "description": "store replicas",
          "type": "integer",
          "format": "int32"
        },
        "trigger_replicas": {
          "description": "trigger replicas",
          "type": "integer",
          "format": "int32"
        },
        "version": {
          "description": "cluster version",
          "type": "string"
        }
      }
    },
    "Connector_info": {
      "description": "Connector info",
      "type": "object",
      "properties": {
        "config": {
          "description": "connector config file",
          "type": "string"
        },
        "kind": {
          "description": "connector kind",
          "type": "string"
        },
        "name": {
          "description": "connector name",
          "type": "string"
        },
        "reason": {
          "description": "connector status reason",
          "type": "string"
        },
        "status": {
          "description": "connector status",
          "type": "string"
        },
        "type": {
          "description": "connector type",
          "type": "string"
        },
        "version": {
          "description": "connector version",
          "type": "string"
        }
      }
    },
    "Controller_info": {
      "description": "controller info",
      "type": "object",
      "properties": {
        "replicas": {
          "description": "controller replicas",
          "type": "integer",
          "format": "int32"
        },
        "storage_size": {
          "description": "controller storage size",
          "type": "string"
        },
        "version": {
          "description": "controller version",
          "type": "string"
        }
      }
    },
    "Gateway_info": {
      "description": "gateway info",
      "type": "object",
      "properties": {
        "version": {
          "description": "gateway version",
          "type": "string"
        }
      }
    },
    "Health_info": {
      "description": "operator info",
      "type": "object",
      "properties": {
        "status": {
          "description": "operator status",
          "type": "string"
        }
      }
    },
    "Store_info": {
      "description": "store info",
      "type": "object",
      "properties": {
        "replicas": {
          "description": "store replicas",
          "type": "integer",
          "format": "int32"
        },
        "storage_size": {
          "description": "store storage size",
          "type": "string"
        },
        "version": {
          "description": "store version",
          "type": "string"
        }
      }
    },
    "Timer_info": {
      "description": "timer info",
      "type": "object",
      "properties": {
        "replicas": {
          "description": "timer replicas",
          "type": "integer",
          "format": "int32"
        },
        "version": {
          "description": "timer version",
          "type": "string"
        }
      }
    },
    "Trigger_info": {
      "description": "trigger info",
      "type": "object",
      "properties": {
        "replicas": {
          "description": "trigger replicas",
          "type": "integer",
          "format": "int32"
        },
        "version": {
          "description": "trigger version",
          "type": "string"
        }
      }
    }
  }
}`))
	FlatSwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "description": "this document describes the functions, parameters and return values of the vanus operator API\n",
    "title": "Vanus Operator Interface Document",
    "version": "v0.2.0"
  },
  "basePath": "/api/v1",
  "paths": {
    "/cluster/": {
      "get": {
        "description": "get Cluster",
        "tags": [
          "cluster"
        ],
        "operationId": "getCluster",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Cluster_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "post": {
        "description": "create Cluster",
        "tags": [
          "cluster"
        ],
        "operationId": "createCluster",
        "parameters": [
          {
            "description": "需要创建的Cluster信息",
            "name": "create",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "$ref": "#/definitions/Cluster_create"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Cluster_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "delete": {
        "description": "delete Cluster",
        "tags": [
          "cluster"
        ],
        "operationId": "deleteCluster",
        "parameters": [
          {
            "description": "delete Cluster",
            "name": "delete",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "$ref": "#/definitions/Cluster_delete"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Cluster_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "patch": {
        "description": "patch Cluster",
        "tags": [
          "cluster"
        ],
        "operationId": "patchCluster",
        "parameters": [
          {
            "description": "patch info",
            "name": "patch",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "$ref": "#/definitions/Cluster_patch"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Cluster_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      }
    },
    "/connector/{name}": {
      "get": {
        "description": "get Connector",
        "tags": [
          "connector"
        ],
        "operationId": "getConnector",
        "parameters": [
          {
            "type": "string",
            "description": "connector name",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Connector_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      }
    },
    "/connectors/": {
      "get": {
        "description": "list Connector",
        "tags": [
          "connector"
        ],
        "operationId": "listConnector",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "$ref": "#/definitions/Connector_info"
                  }
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "post": {
        "description": "create Connector",
        "tags": [
          "connector"
        ],
        "operationId": "createConnector",
        "parameters": [
          {
            "description": "Connector info",
            "name": "connector",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "$ref": "#/definitions/Connector_info"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Connector_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      },
      "delete": {
        "description": "delete Connector",
        "tags": [
          "connector"
        ],
        "operationId": "deleteConnector",
        "parameters": [
          {
            "type": "string",
            "description": "connector name",
            "name": "name",
            "in": "query"
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Connector_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      }
    },
    "/healthz/": {
      "get": {
        "description": "healthz",
        "tags": [
          "healthz"
        ],
        "operationId": "healthz",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "type": "object",
              "required": [
                "code",
                "data",
                "message"
              ],
              "properties": {
                "code": {
                  "type": "integer",
                  "format": "int32",
                  "default": 200
                },
                "data": {
                  "type": "object",
                  "$ref": "#/definitions/Health_info"
                },
                "message": {
                  "type": "string",
                  "default": "success"
                }
              }
            }
          }
        }
      }
    }
  },
  "definitions": {
    "APIResponse": {
      "type": "object",
      "properties": {
        "code": {
          "type": "integer",
          "format": "int32"
        },
        "data": {
          "type": "object"
        },
        "message": {
          "type": "string"
        }
      }
    },
    "Cluster_create": {
      "description": "cluster create params",
      "type": "object",
      "properties": {
        "controller_replicas": {
          "description": "controller replicas",
          "type": "integer",
          "format": "int32"
        },
        "controller_storage_size": {
          "description": "controller storage size",
          "type": "string"
        },
        "store_replicas": {
          "description": "store replicas",
          "type": "integer",
          "format": "int32"
        },
        "store_storage_size": {
          "description": "store storage size",
          "type": "string"
        },
        "version": {
          "description": "cluster version",
          "type": "string"
        }
      }
    },
    "Cluster_delete": {
      "description": "cluster create params",
      "type": "object",
      "properties": {
        "force": {
          "description": "true means force delete, default false",
          "type": "boolean",
          "default": false
        }
      }
    },
    "Cluster_info": {
      "description": "cluster info",
      "type": "object",
      "properties": {
        "status": {
          "description": "cluster status",
          "type": "string"
        },
        "version": {
          "description": "cluster version",
          "type": "string"
        }
      }
    },
    "Cluster_patch": {
      "description": "cluster patch params",
      "type": "object",
      "properties": {
        "controller_replicas": {
          "description": "controller replicas",
          "type": "integer",
          "format": "int32"
        },
        "store_replicas": {
          "description": "store replicas",
          "type": "integer",
          "format": "int32"
        },
        "trigger_replicas": {
          "description": "trigger replicas",
          "type": "integer",
          "format": "int32"
        },
        "version": {
          "description": "cluster version",
          "type": "string"
        }
      }
    },
    "Connector_info": {
      "description": "Connector info",
      "type": "object",
      "properties": {
        "config": {
          "description": "connector config file",
          "type": "string"
        },
        "kind": {
          "description": "connector kind",
          "type": "string"
        },
        "name": {
          "description": "connector name",
          "type": "string"
        },
        "reason": {
          "description": "connector status reason",
          "type": "string"
        },
        "status": {
          "description": "connector status",
          "type": "string"
        },
        "type": {
          "description": "connector type",
          "type": "string"
        },
        "version": {
          "description": "connector version",
          "type": "string"
        }
      }
    },
    "Controller_info": {
      "description": "controller info",
      "type": "object",
      "properties": {
        "replicas": {
          "description": "controller replicas",
          "type": "integer",
          "format": "int32"
        },
        "storage_size": {
          "description": "controller storage size",
          "type": "string"
        },
        "version": {
          "description": "controller version",
          "type": "string"
        }
      }
    },
    "Gateway_info": {
      "description": "gateway info",
      "type": "object",
      "properties": {
        "version": {
          "description": "gateway version",
          "type": "string"
        }
      }
    },
    "Health_info": {
      "description": "operator info",
      "type": "object",
      "properties": {
        "status": {
          "description": "operator status",
          "type": "string"
        }
      }
    },
    "Store_info": {
      "description": "store info",
      "type": "object",
      "properties": {
        "replicas": {
          "description": "store replicas",
          "type": "integer",
          "format": "int32"
        },
        "storage_size": {
          "description": "store storage size",
          "type": "string"
        },
        "version": {
          "description": "store version",
          "type": "string"
        }
      }
    },
    "Timer_info": {
      "description": "timer info",
      "type": "object",
      "properties": {
        "replicas": {
          "description": "timer replicas",
          "type": "integer",
          "format": "int32"
        },
        "version": {
          "description": "timer version",
          "type": "string"
        }
      }
    },
    "Trigger_info": {
      "description": "trigger info",
      "type": "object",
      "properties": {
        "replicas": {
          "description": "trigger replicas",
          "type": "integer",
          "format": "int32"
        },
        "version": {
          "description": "trigger version",
          "type": "string"
        }
      }
    }
  }
}`))
}
