# Welcome to e2eChat

A lightweight real time chat application built with Go, WebSockets, and Svelte.

***

## Tech Stack

* [Go](https://go.dev): Used for the backend chat server and CLI client

* [coder/websocket](https://github.com/coder/websocket): A minimal WebSocket library for Go

* [SvelteKit](https://kit.svelte.dev): Used for the frontend chat interface

* [TypeScript](https://www.typescriptlang.org): Adds type safety to the frontend

* HTML/CSS, JavaScript

***

## Getting Started

### 1. Clone Project

> git clone https://github.com/eruigu/e2eChat.git

### 2. Install root dependencies

> npm install

### 3. Install frontend dependencies

> cd frontend  
> npm install

### 4. Run the full app

From the project root:

> npm run dev

This starts:

> the Go backend on `localhost:8080`  
> the Svelte frontend with Vite

### 5. Run backend only

> npm run backend

### 6. Run frontend only

> npm run frontend

***

## Project Structure

```bash
e2eChat/
├── backend/
│   ├── chat.go
│   ├── client.go
│   ├── main.go
│   ├── server.go
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   ├── static/
│   ├── package.json
│   └── svelte.config.js
├── package.json
└── package-lock.json
