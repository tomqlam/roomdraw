# Backend Setup Guide

This guide will walk you through setting up and running the RoomDraw backend.

## Prerequisites

Before getting started, make sure you have the following installed on your machine:

- Docker

## Running the Backend

To run the RoomDraw backend, follow these steps:

1. Open your terminal or command prompt.

2. Navigate to the directory where you have the backend code.

3. Run the following command:

    ```bash
    docker run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
    ```

    This command will start a Docker container with the RoomDraw backend image. It will map port 8080 on your local machine to port 8080 in the container and mount the current directory as `/app` inside the container.

4. Once the container is running, you can access the backend API at `http://localhost:8080`.

That's it! You have successfully set up and run the RoomDraw backend.

## Additional Information

- If you encounter any issues or have questions, please refer to the [backend documentation](/path/to/backend/documentation).

- For more advanced configuration options, please consult the [backend configuration guide](/path/to/backend/configuration).
