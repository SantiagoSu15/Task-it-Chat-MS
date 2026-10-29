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
- `internal/infra/inbound/http`: adaptador de entrada HTTP (Gin).
- `internal/infra/inbound/websocket`: adaptador de entrada WebSocket.
- `internal/infra/outbound/websocket`: Hub y adaptador de distribución WebSocket.
- `internal/infrastructure/config`: carga de configuración.
- `internal/infrastructure/redis`: wiring técnico de cliente Redis.
- `internal/infrastructure/rabbitmq`: wiring técnico de cliente RabbitMQ.
- `deployments`: artefactos de despliegue.
- `docs`: documentación adicional.

## WebSocket y persistencia Redis

El servidor escucha en `HTTP_PORT` (por defecto `8080`) y expone `GET /ws/chat`. Las conexiones y salas activas viven en memoria; Redis conserva los chats y mensajes, pero no entrega mensajes nuevos a clientes desconectados.

Los chats y mensajes se almacenan en Redis sin TTL. `REDIS_ADDR`, `REDIS_PASSWORD` y `REDIS_DB` configuran la conexión. El historial se guarda en un `HASH` por chat y su orden en un `ZSET`; los índices de participantes y tareas son `SET` persistentes. El borrado explícito elimina el chat, sus mensajes y sus índices.

El protocolo usa JSON:

```json
{"type":"chat:EnterRoom","chatId":"chat-123"}
{"type":"chat:sendMessage","chatId":"chat-123","content":"Hola","messageId":"msg-456"}
```

Los mensajes emitidos tienen la forma:

```json
{"type":"chat:newMessage","data":{"messageId":"msg-456","chatId":"chat-123","senderId":123,"content":"Hola","timestamp":"2026-10-28T12:00:00Z"}}
```

Para dejar preparado el límite de autenticación, el adaptador acepta temporalmente `X-User-ID` en el handshake y lo usa como identidad del contexto; este header no constituye autenticación y debe ser reemplazado por middleware autenticado antes de producción. El cliente no puede enviar `senderId` en el evento.

Con dos clientes, por ejemplo `websocat`:

```sh
websocat -H='X-User-ID: 101' ws://localhost:8080/ws/chat
websocat -H='X-User-ID: 202' ws://localhost:8080/ws/chat
```

En ambos clientes enviar primero `chat:EnterRoom` para `chat-123` y después enviar el evento `chat:sendMessage` desde uno de ellos. Solo los clientes de esa sala reciben `chat:newMessage`.
