# GOE

GOE is a simple and easy to use web development framework for Go. It is designed to be simple and easy to use, and to be
able to quickly build a web application. It is based on the [GoFiber](https://gofiber.io/) framework, and adds some
useful features.

GOE learned from practices of Golang projects, Java web development frameworks, such as Spring Boot, and tried to
provide a similar
experience but simpler and lighter.

Thanks for the following projects that GOE relies on or inspired by:

- [GoFiber](https://gofiber.io/) - For handling HTTP related tasks.
- [GoFr](https://gofr.dev/) - For the project structure and interface design.
- [Qmgo](https://github.com/qiniu/qmgo) - For the MongoDB operations.
- [Gookit Validate](https://github.com/gookit/validate) - For the data validation.
- [PocketBase](https://pocketbase.io/) - For the mailer implementation and interface design.
- [Delayqueue](https://github.com/HDT3213/delayqueue) - For the message queue implementation.
- [Zap](https://github.com/uber-go/zap) - For the logger implementation.
- [Kelindar Event](https://github.com/kelindar/event) - For the event bus implementation.

## Goals

The only goal of GOE is to provide a simple and easy to use web development framework for Go. Developers should only
focus on the business logic, and GOE will handle the rest.

## Supported Database

Currently, GOE only supports MongoDB. Will support more SQL databases in the future.

## Plans

- [ ] Code Generator, to generate the project structure and code.
- [ ] gRPC Support, based on Buf.