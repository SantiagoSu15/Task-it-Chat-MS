# Task-it-Chat-MS

Scaffolding inicial del microservicio de chat en Go con arquitectura hexagonal (Ports & Adapters).

## Estructura

- `cmd/chat`: punto de entrada de la aplicación.
- `internal/bootstrap`: preparación del arranque y ensamblado de componentes.
- `internal/domain`: núcleo del dominio (entidades y reglas de negocio).
- `internal/application/port/inbound`: puertos de entrada para casos de uso.
- `internal/application/port/outbound`: puertos de salida para infraestructura externa.
- `internal/application/service`: servicios de aplicación/casos de uso.
- `internal/adapter/inbound/http`: adaptadores de entrada HTTP (Gin).
- `internal/adapter/inbound/websocket`: adaptadores de entrada WebSocket.
- `internal/adapter/inbound/broker`: adaptadores de entrada para eventos de broker.
- `internal/adapter/outbound/redis`: adaptadores de salida para Redis.
- `internal/adapter/outbound/rabbitmq`: adaptadores de salida para RabbitMQ.
- `internal/infrastructure/config`: carga de configuración.
- `internal/infrastructure/redis`: wiring técnico de cliente Redis.
- `internal/infrastructure/rabbitmq`: wiring técnico de cliente RabbitMQ.
- `deployments`: artefactos de despliegue.
- `docs`: documentación adicional.

## Estado

Este repositorio contiene únicamente el esqueleto base, sin implementación de lógica de negocio, endpoints, consumidores, repositorios o conexiones externas.
