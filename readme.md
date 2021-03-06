<h1>My Snippetbox implementation From <a href ="https://lets-go-further.alexedwards.net/">Lets Go<a> Book</h1>

<div>
<img src="https://github.com/Xyedo/snippetbox/blob/main/image/first.jpg" width="350" alt="homepage"/>
<img src="https://github.com/Xyedo/snippetbox/blob/main/image/after-login.jpg" width="350" alt="after login"/>
<img src="https://github.com/Xyedo/snippetbox/blob/main/image/account.jpg" width="350" alt="after login"/>
<img src="https://github.com/Xyedo/snippetbox/blob/main/image/change%20password.jpg" width="350" alt="after login"/>
</div>

### Using Go 1.18 ![check-code-coverage](https://img.shields.io/badge/coverage-72%25-orange)

<h2>My Change</h2>
<ul>
<li>Added Response Time & status, kinda like morgan in express</li>
<li>Will not Change Middleware to compossable middleware like express because go middleware doesnt behave like express</li>
</ul>
<h2>Finished - What I Learn</h2>
<ul>
<li>Best Practices building web app using go</li>
<li>using std package like http, flag, context, httptest, and many more</li>
<li>unit testing, mocking and end-to-end testing, integration testing in GO </li>
</ul>
<h2>TODO List Exercises </h2>

- [x] Add About Page to the App
- [x] Add a debug mode
- [x] more http e2e testing
- [x] add Account page to the app
- [x] Redirect appropriately after login
- [x] Implement Change Password Features 

## Installation

<details>
  <summary>Pre-Installation</summary>

  1. Having MySQL install
  2. creating new user, snippetbox db, users table, and snippets table
  ```sql
  mysql -u root -p
  #enter your password
  
  CREATE DATABASE snippetbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
  
  USE snippetbox;
  CREATE TABLE snippets (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created DATETIME NOT NULL,
    expires DATETIME NOT NULL
  );
  
  CREATE INDEX idx_snippets_created ON snippets(created);
  
  CREATE USER 'web'@'localhost';
  GRANT SELECT, INSERT, UPDATE, DELETE ON snippetbox.* TO 'web'@'localhost';
  ALTER USER 'web'@'localhost' IDENTIFIED BY 'pass';
  
  CREATE TABLE users (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created DATETIME NOT NULL
   );
   
   ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);
  ```
  
</details>
You can install the project by forking or cloning
You need to add s self-signed TLS certificate

```bash
mkdir project_path/tls
cd project_path/tls
go run /usr/local/go/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=lslocalhost
```

## Running the Project
you can run the project by :
```bash
go run ./cmd/web #Check https://localhost:4000 for the web
```

you can run the test by :

```bash
go test -v ./...
```
you can run the coverage by :

```bash
go test -cover ./...
```

