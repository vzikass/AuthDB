# AuthDB
![GitHub contributors](https://img.shields.io/github/contributors/vzikass/AuthDB)
![GitHub last commit](https://img.shields.io/github/last-commit/vzikass/AuthDB)
![Docker Image Version](https://img.shields.io/docker/v/_/alpine)
![GitHub Created At](https://img.shields.io/github/created-at/vzikass/AuthDB)
![Go version](https://img.shields.io/github/go-mod/go-version/vzikass/AuthDB)


**AuthDB** is a project I wrote to practice the following things:
+ User registration with automatic addition of the user to the database.
  + The password is automatically hashed and the database receives the already hashed password.
+ User authentication by [JWT](https://jwt.io/introduction) (json web token).
+ Creating a cookie for the user.
+ HTTP request methods.
+ Querying the database.
+ Docker image and containers.
+ Kubernetes([minikube](https://kubernetes.io/docs/tutorials/hello-minikube/))
+ CI/CD (Github Actions)
+ Testing
  
And other things I've encountered in the writing process.

*This list will grow as I apply what I've learned to it.* 

[Here is the answer to your probably first question when you see this project](https://github.com/vzikass/AuthDB?tab=readme-ov-file#why-did-i-develop-this-project)

## Project launch
### Requirements
**The golang 1.22.1 version is required to install and run the project**

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
*I don't see much point in testing, because in the process I gradually test the work myself. But since the project is intended for personal use, you can practice and make some Unit Tests, Integration Test or other tests. Perhaps I will add more tests if I think it's necessary.*

## Achievements
- [X] Write terrible code
- [ ] Get the job done
- [X] Sitting for days on a single bug
## Technology stack:
* DB: _Postgres_ ([pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v4/pgxpool))
* Containers: _Docker_, _Docker-compose_
* Front: [_bootstrap_](https://getbootstrap.com/), _html_, _css_, _js_ scripts
* And probabbly something else I forgot
  
## Why did I develop this project?  
~~Just to be~~

## Project team
1. Вячеслав Ивкин (Vyacheslav Ivkin) https://github.com/vzikass

## Contributing
Bug reports and/or pull requests are welcome!\
*I leave here my contact in telegram for communication(ru, en) - **click :point_right: [tg](https://t.me/vzikass)***

## License
The module is available as open source under the terms of the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)


>As we can see, perfection is achieved not when there is nothing more to add, but when nothing more can be taken away.
>
> \- Antoine de Saint-Exupéry
