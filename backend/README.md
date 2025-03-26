# Backend Setup Guide

This guide will walk you through setting up and running the RoomDraw backend.

## Prerequisites

Before getting started, make sure you have the following installed on your machine:

- Docker

## Running the Backend

To run the RoomDraw backend, follow these steps:

1. Open your terminal or command prompt.

2. Navigate to the directory where you have the backend code.

3. Choose your environment:

### Development Environment

For local development, run:
```bash
docker build -t roomdraw-backend .
docker run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
```

### Production Environment

For production deployment, run:
```bash
podman build -t roomdraw-backend --build-arg ENV=production .
podman run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
```
<!-- 
If you're running podman on a server on which you need root privs, run the above command locally to build the image. Then do:

1. ```bash
    docker save roomdraw-backend > /path/to/roomdraw-backend.tar
    ```
2. ```bash
    scp /path/to/roomdraw-backend.tar user@destination_host:path_on_destination 
    ```
3. ```bash
    podman load < roomdraw-backend.tar
    ```
4. ```bash
    podman run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
    ``` -->

<!-- 4. Once the container is running, you can access the backend API at `http://localhost:8080`. -->

## Environment Variables

The backend uses different environment variables for development and production:

- `.env.development`: Contains local development settings
- `.env.production`: Contains production deployment settings

Make sure to update these files with the appropriate values for your environment.

That's it! You have successfully set up and run the RoomDraw backend.
