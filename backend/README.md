# Backend Setup Guide

This guide will walk you through setting up and running the RoomDraw backend.

## Prerequisites

Before getting started, make sure you have the following installed on your machine:

- Docker

## Running the Backend

To run the RoomDraw backend, follow these steps:

1. Open your terminal or command prompt.

2. Navigate to the directory where you have the backend code.

3. Run the following commands:
    Build the Docker image from the Dockerfile in the current directory and tag the image as `roomdraw-backend`.
    ```bash
    docker build -t roomdraw-backend .
    ```

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

    Run the following command to start the Docker container:
    ```bash
    docker run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
    ```

    Or if you're using podman:

    ```bash
    podman run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
    ```

    This command will start a Docker container with the RoomDraw backend image. It will map port 8080 on your local machine to port 8080 in the container and mount the current directory as `/app` inside the container.

4. Once the container is running, you can access the backend API at `http://localhost:8080`.

That's it! You have successfully set up and run the RoomDraw backend.

## Additional Information

- If you encounter any issues or have questions, please refer to the [backend documentation](/path/to/backend/documentation).

- For more advanced configuration options, please consult the [backend configuration guide](/path/to/backend/configuration).
