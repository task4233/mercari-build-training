# Database
## 1. Write into a database

```sql
CREATE TABLE items (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    category VARCHAR(255),
    image_name VARCHAR(255)
);
```

```bash
$ sqlite3 db/mercari.sqlite3 < db/items.db
$ sqlite3 ./db/mercari.sqlite3 
sqlite> .schema items
CREATE TABLE items (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    category VARCHAR(255),
    image_name VARCHAR(255)
);
```

---

> Change the endpoints `GET /items` and `POST /items` such that items are saved into the database and can be returned based on GET request.

### GET /items
- connect to the SQLite3 database
- invoke SQL to collect all of items
- return them to a client

### POST /items
- connect to the SQLite3 database
- invoke SQL to insert a record given from a client

### Initialization
To prepare an initial data, invoke the following SQL:

```bash
$ sqlite3 ../db/mercari.sqlite3
sqlite> INSERT INTO items (name, category, image_name) VALUES ("jacket", "fashion", "510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg");
```

### Self-QA
```bash
$ # for GET /items
$ curl localhost:9000/items
{"items":[{"name":"jacket","category":"fashion","image_name":"510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg"}]}
$ # for POST /items
$ curl -X POST -F "name=shoes" http://localhost:9000/items
{"message":"item received: shoes"}
$ curl localhost:9000/items
{"items":[{"name":"jacket","category":"fashion","image_name":"510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg"},{"name":"shoes","category":"unknown","image_name":"default.jpg"}]}
```

---

> What are the advantages of saving into a database such as SQLite instead of saving into a single JSON file?

- to manage structured data
- to manage data safely with multiple people
- to operate data declaratively

## 2. Search for an item

> Make an endpoint to return a list of items that include a specified keyword called GET /search.

```bash
# Request a list of items containing string "jacket"
$ curl -X GET 'http://127.0.0.1:9000/search?keyword=jacket'
# Expected response for a list of items with name containing "jacket"
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```

---

- goal
  - search items by the `keyword` value given the query in URL
- requirements
  - get `keyword` value from the query
  - enumerate items by filtering with the `keyword` value
  - return them to the client

### Self-QA

```bash
$ curl -X GET 'http://127.0.0.1:9000/search?keyword=jacket'
{"items":[{"name":"jacket","category":"fashion","image_name":"510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg"}]}
$ curl -X GET 'http://127.0.0.1:9000/search?keyword='      
{"items":[{"name":"jacket","category":"fashion","image_name":"510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg"},{"name":"shoes","category":"unknown","image_name":"default.jpg"}]}
$ curl -X GET 'http://127.0.0.1:9000/search?keyword=h'
{"items":[{"name":"shoes","category":"unknown","image_name":"default.jpg"}]}
```

