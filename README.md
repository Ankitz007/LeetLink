# LeetLink

> [!NOTE]  
> This is a learning project and is not intended for any other use. Please use it only for educational purposes.

LeetLink is a URL redirection service that simplifies access to LeetCode problem descriptions. It leverages a Redis database to store and retrieve problem slugs based on problem IDs.

## Features

- **URL Redirection**: Seamlessly redirects users to LeetCode problem description pages using problem IDs.
- **Redis Integration**: Efficient storage and retrieval of problem slugs with Redis.
- **Automated Updates**: Periodically fetches and updates problem data from LeetCode's API using a cron job.
- **Live Development**: Supports live reloading during development with `air`.
- **Dockerized Deployment**: Easily deployable with Docker and Docker Compose.

## Setup Instructions

### Prerequisites

- **Docker** and **Docker Compose** installed.
- **Go** (version 1.23 or higher) installed if running locally without Docker.
- **Redis** server running and accessible.

### Running Locally with Docker

1. Clone the repository:

    ```bash
    git clone https://github.com/Ankitz007/leetlink.git
    cd leetlink
    ```

2. Create a `.env` file by copying the example:

    ```bash
    cp .env.example .env
    ```

3. Update the `.env` file with your Redis credentials and a secret for the cron job.

4. Build and start the Docker container:

    ```bash
    docker-compose up -d leetlink
    ```

5. Access the application at [http://localhost:8080](http://localhost:8080).

## API Endpoints

### `/api/leetlink/`

- **Method**: `GET`
- **Query Parameter**: `problem_id` (required)
- **Description**: Redirects to the LeetCode problem description page based on the provided problem ID.

### `/api/cron/`

- **Method**: `POST`
- **Headers**: `Authorization: Bearer <CRON_SECRET>`
- **Description**: Fetches and updates problem data from LeetCode's API.

## Deployment

LeetLink can be deployed using **Vercel**. The `vercel.json` file is pre-configured to handle API routes and serve a custom 404 page.

### Cron Job Configuration

The `vercel.json` file includes a cron job configuration to run the `/api/cron` endpoint daily at 12:50 PM UTC. Ensure the `CRON_SECRET` in your `.env` file matches the secret used in the cron job.
