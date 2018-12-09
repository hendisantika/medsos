# medsos

Social Medial Activity Feed

## API Endpoint

| Endpoint  | Method | Request           | Response    | Note               |
|-----------|--------|-------------------|-------------|--------------------|
| /register | POST   | { actor: "irvan", | 201 Created | register new actor |
|           |        |   friends: [] }   |             |                    |
|           |        |                   |             |                    |
| /follow/{actor}/ | POST | { actor: "niko" } | 204 No Content | follow friend |
| /follow/{actor}/{friend} | DELETE | | 204 No Content | unfollow friend |
| /feeds | POST | {actor: ..., verb: ..., object: ..., target: ...}| 200 OK | New activity feed |
| /feeds/{actor}/ | GET | | 200 OK | get activity feed |

## Usage

1. Edit configuration in `.env`

2. run `docker-compose up` inside project directory to compile, run API mongo and mongo service.

3. Access API server using web client to address `http://localhost:8080` or `MEDSOS_ADDRESS` from `.env`

## Example

1. Register new actor
    ```bash
    curl -v X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"actor":"irvan","friends":[]}'
    ```

2. Post new feed
    ```bash
    curl -v -X POST http://localhost:8080/feeds -H "Content-type: application/json" -d '{"actor":"irvan","verb":"post","object":"post:1"}'
    ```

3. Get feeds
    ```bash
     curl -v -X GET http://localhost:8080/feeds/irvan/
    ```
 
4. Follow friend
    ```bash
    curl -v -X POST http://localhost:8080/follow/irvan/ -H "Content-type: application/json" -d '{"actor":"niko"}'
    ```

5. Unfollow friend
    ```bash
    curl -v -X DELETE http://localhost:8080/follow/irvan/niko
    ```