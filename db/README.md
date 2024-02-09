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
