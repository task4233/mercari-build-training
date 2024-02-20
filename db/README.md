# Database(DB)
What is "Database"?

According to the article by Oracle, "Database" is explained as follows:

> A database is an organized collection of structured information, or data, typically stored electronically in a computer system. A database is usually controlled by a database management system (DBMS). Together, the data and the DBMS, along with the applications that are associated with them, are referred to as a database system, often shortened to just database.

> Data within the most common types of databases in operation today is typically modeled in rows and columns in a series of tables to make processing and data querying efficient. The data can then be easily accessed, managed, modified, updated, controlled, and organized. Most databases use structured query language (SQL) for writing and querying data.

There are various types of DBs such as MySQL, PostgreSQL, SQLite, Oracle Database, Spanner, SpiceDB and etc... They have different features and different use cases in each. Their details are omitted because it is too much information to explain them enough time. I think best way to learn "Database" is to try to design and use like "Nothing beats experience". but for someone who is interested in databases, please let me share some recommendation from me :+1:

ref:
- [[Performance] SQL Antipatterns](https://www.oreilly.com/library/view/sql-antipatterns/9781680500073/)
  - This book introduces some anti-pattern to make our SQL which is the langueage to handle DB slow and the way to resolve them.
- [[Mechanism] Database Internals](https://www.oreilly.com/library/view/database-internals/9781492040330/)
  - This book provides information about internal mechanisms of DBMS and distributed systems.
- [[Theory] Foundations of Databases](http://webdam.inria.fr/Alice/)
  - This book provides detailed information on database fundamentals and theory.

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

## 3. Move the category information to a separated table

> Modify the database as follows. That makes it possible to change the category names without modifying the all categories of items in the items table. Since GET items should return the category name as before, join these two tables when returning responses.

**items table**

| id   | name   | category_id | image_filename                                                       |
| :--- | :----- | :---------- | :------------------------------------------------------------------- |
| 1    | jacket | 1           | 510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg |
| 2    | ...    |             |                                                                      |

**category table**

| id   | name    |
| :--- | :------ |
| 1    | fashion |
| ...  |         |


---

- goal
  - separate table
- requirements
  - create a new table `category`
  - update the `items` table

SQLite does not provide the feature to change the type of columns which has already defined. So, please let me defining from the beginning.

```sql
CREATE TABLE category (
  id INT PRIMARY KEY,
  name VARCHAR(255)
);

CREATE TABLE items (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    category_id int,
    image_name VARCHAR(255),
    foreign key (category_id) references category(id)
);
```

Apply changes as follows.

```bash
$ # remove the existing DB
$ rm db/mercari.sqlite3 
$ # construct new definitions
$ sqlite3 db/mercari.sqlte3
sqlite> CREATE TABLE category (
  id INT PRIMARY KEY,
  name VARCHAR(255)
);
sqlite> .schema category
CREATE TABLE category (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255)
);
sqlite> CREATE TABLE items (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    category_id int,
    image_name VARCHAR(255),
    foreign key (category_id) references category(id)
);
sqlite> .schema items
CREATE TABLE items (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    category_id int,
    image_name VARCHAR(255),
    foreign key (category_id) references category(id)
);
```

Add values represented in the example into DB as follows:

```bash
INSERT INTO category (id, name) VALUES (1, "fashion");
INSERT INTO items (id, name, category_id, image_name) VALUES (1, "jacket", 1, "510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg");
```

Confirm the inserted value.

```bash
sqlite> SELECT * FROM items;
1|jacket|1|510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg
sqlite> SELECT * FROM category;
1|fashion
sqlite> SELECT i.id, i.name, c.name, i.image_name FROM items as i JOIN category as c ON i.category_id=c.id;
1|jacket|fashion|510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg
```

---

> What is database **normalization**?

In summary, the database normalization is to divide large tables into smaller, more manageable tables and define relationships between them. The purpose is to reduce redundancy.

By the way, what are "redundancy"?

"High redundancy" means that data with the same meaning is stored scatterly. Let's say, for example, imagine a case of weight measurement. There're both weight(kg) and weight(g) are stored. This implies the same data, but in different units. Their data can be derived each other. Then, the condition is called "high redundancy".

One of the pros is that by reducing redundancy, the risk of data inconsistencies is reduced. For example, in the previous example of weight, let's say the measurements were incorrect and required to be updated. Then, suppose you updated the information for weight(kg), but forgot to update in weight(g). This would cause inconsistencies in the information, and other people would not know which information was correct when they looked at them. The normalisation can help to avoid this issue.

One of the cons is time consuming when collecting the data may increase because the information is distributed. An analogy would be like data being written in various notebooks; it is likely to be more difficult to find information from several notebooks together than from one notebook.

Therefore, appropriate normalisation is needed to be implemented, taking into account the pros and cons.
