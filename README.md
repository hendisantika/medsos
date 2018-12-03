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
| /feeds/{actor} | GET | | 200 OK | get activity feed |

## Usage

1. Edit configuration in `.env`

2. run `docker-compose up` inside project directory to compile, run API mongo and mongo service.

3. Access API server using web client to address `http://localhost:8080` or `MEDSOS_ADDRESS` from `.env`
