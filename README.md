# AuthDB
![GitHub contributors](https://img.shields.io/github/contributors/vzikass/AuthDB)
![GitHub last commit](https://img.shields.io/github/last-commit/vzikass/AuthDB)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/vzikass/AuthDB/ci.yml)
![Docker Image Version](https://img.shields.io/docker/v/_/alpine)
![GitHub Created At](https://img.shields.io/github/created-at/vzikass/AuthDB)
![Go version](https://img.shields.io/github/go-mod/go-version/vzikass/AuthDB)

## Preview
**Login Page** (http://localhost:4444/login)

![Preview of the login page](/public/jpg/login.png)

**SignUp Page** (http://localhost:4444/signup)
![Preview of the SignUP page](/public/jpg/signup.png)

**Main Page** (http://localhost:4444 after registration)
![Preview of the Main page](/public/jpg/main.png)
_logout, delete account and update data, these buttons work and perform their functionality_

_also if you go to http://localhost:4444/users you can see your account and other users_

-----------

**AuthDB** is a project I wrote to practice the following things:
+ User registration with automatic addition of the user to the database
  + The password is automatically hashed and the database receives the already hashed password
+ User authentication by [JWT](https://jwt.io/introduction)
+ Creating a cookie for the user
+ HTTP request methods
+ Querying the database
+ Docker image and containers
+ CI/CD (Github Actions)
+ Git/Github
+ Postman
+ Markdown
+ Testing (with database for tests)
+ Kafka (simple configuration)
+ Mocks ([sarama](https://github.com/IBM/sarama))
  
And other things I've encountered in the writing process.

*This list will grow as I apply what I've learned to it.* 

## Project launch
### Requirements
**The golang 1.22 version is required to install and run the project**

**Make sure you have all the necessary dependencies installed beforehand**

```
git clone https://github.com/vzikass/AuthDB
```
When you're ready, start your application by running:
```
docker compose up --build
```
***Wait for the docker container to load***

Application will be available at [localhost:4444](http://localhost:4444)

## Testing
*Added testing of some functionality*\
You can check it :point_right: [here](https://github.com/vzikass/AuthDB/blob/main/authdb_test.go) :point_left:


## Achievements
- [X] Write terrible code
- [ ] Get the job done
- [X] Sitting for days on a single bug
## Technology stack:
* *Golang* 1.22 
* Git
* Web server: [nginx](https://nginx.org/en/)
* Broker: [Apache Kafka](https://kafka.apache.org/)
* DB: _Postgres_ ([pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v4/pgxpool))
* Containers: _Docker_, _Docker-compose_
* Front: [_bootstrap_](https://getbootstrap.com/), _html_ (some helpful youtube videos i used: [1](https://www.youtube.com/watch?v=hlwlM4a5rxg), [2](https://www.youtube.com/watch?v=EzXdxvO1htA&t=672s)), _css_
* CI/CD ([Github Actions](https://docs.github.com/en/actions))
* Postman
* And probabbly something else I forgot
  
## Why did I develop this project?  
~~Just to be~~

## Project team
1. Вячеслав Ивкин (Vyacheslav Ivkin) https://github.com/vzikass

## Contributing
Bug reports and/or pull requests are welcome!\
*I leave here my contact in telegram for communication(ru, en) - :point_right: [tg](https://t.me/vzikass)

## License
The module is available as open source under the terms of the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)


>As we can see, perfection is achieved not when there is nothing more to add, but when nothing more can be taken away.
>
> \- Antoine de Saint-Exupéry
