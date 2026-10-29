# Task-it-Chat-MS

Scaffolding inicial del microservicio de chat en Go con arquitectura hexagonal (Ports & Adapters).

## Estructura

- `cmd/chat`: punto de entrada de la aplicación.
- `internal/bootstrap`: preparación del arranque y ensamblado de componentes.
- `internal/domain`: núcleo del dominio (entidades y reglas de negocio).
- `internal/domain/model`: modelos `Chat`, `Message` y `User`.
- `internal/domain/repository`: contratos de persistencia del dominio.
- `internal/application/port/inbound`: puertos de entrada para casos de uso.
- `internal/application/port/outbound`: puertos de salida para infraestructura externa.
- `internal/application/service`: servicios de aplicación/casos de uso.
- `internal/infra/inbound/http`: adaptador de entrada HTTP (Chi).
- `internal/infra/inbound/websocket`: esqueleto del adaptador de entrada WebSocket.
- `internal/infra/outbound/websocket`: contrato mínimo del adaptador de salida WebSocket.
- `internal/infrastructure/config`: carga de configuración.
- `internal/infrastructure/redis`: wiring técnico de cliente Redis.
- `internal/infrastructure/rabbitmq`: wiring técnico de cliente RabbitMQ.
- `deployments`: artefactos de despliegue.
- `docs`: documentación adicional.

## Estado

El repositorio contiene los contratos y servicios base del dominio de chat, además del endpoint de historial. No incluye persistencia, consumidores RabbitMQ, conexiones Redis ni un Hub WebSocket real.
