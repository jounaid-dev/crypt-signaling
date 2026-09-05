# CRYPT Signaling Server

WebSocket signaling server for the CRYPT secure communication application.

## Purpose

This server provides the signaling layer required to establish peer-to-peer connections between CRYPT users.

It routes:

- WebRTC offers
- WebRTC answers
- ICE candidates

The server does not decrypt encrypted application payloads.

## Requirements

- Go 1.26.5 or compatible Go version
- Gorilla WebSocket v1.5.3

## Project Structure

crypt-signaling/
├── main.go
├── go.mod
└── README.md
