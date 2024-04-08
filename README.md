# otus-task-4  <br/>
  <br/>
clone https://github.com/xost/otus-task-4.git  <br/>
  <br/>
cd otus-task-4  <br/>
  <br/>
helm install usersapp users/.  <br/>
  <br/>
CREATE:  <br/>
in:  <br/>
    curl -X POST http://arch.homework/api/users -H "Content-Type: application/json" -d '{"name":"Ivan","email":"ivan@mail.ru"}'  <br/>
out:  <br/>
    User with email=ivan@mail.ru was created  <br/>
in:  <br/>
    curl -X POST http://arch.homework/api/users -H "Content-Type: application/json" -d '{"name":"Semen","email":"semen@mail.ru"}'  <br/>
out:  <br/>
    User with email=semen@mail.ru was created  <br/>
in:  <br/>
    curl -X POST http://arch.homework/api/users -H "Content-Type: application/json" -d '{"name":"Gagarin","email":"gagarin@mail.ru"}'  <br/>
out:  <br/>
    curl -X POST http://arch.homework/api/users -H "Content-Type: application/json" -d '{"name":"Gagarin","email":"gagarin@mail.ru"}'  <br/>
  <br/>
GET USER LIST:  <br/>
in:  <br/>
    curl -X GET http://arch.homework/api/users  <br/>
out:  <br/>
    [{"id":1,"email":"ivan@mail.ru","name":"Ivan"},{"id":2,"email":"semen@mail.ru","name":"Semen"},{"id":3,"email":"gagarin@mail.ru","name":"Gagarin"}]  <br/>
  <br/>
GET USER:  <br/>
in:  <br/>
    curl -X GET http://arch.homework/api/users/2  <br/>
out:  <br/>
    {"id":2,"email":"semen@mail.ru","name":"Semen"}  <br/>
  <br/>
UPDATE:  <br/>
in:  <br/>
    curl -X PUT http://arch.homework/api/users/1 -H "Content-Type: application/json" -d '{"name":"Vanya","email":"ivan@mail.ru"}'  <br/>
out:  <br/>
    Updated user with id=1  <br/>
  <br/>
  <br/>
GET USER:  <br/>
in:  <br/>
    curl -X GET http://arch.homework/api/users/1  <br/>
out:  <br/>
    {"id":1,"email":"ivan@mail.ru","name":"Vanya"}  <br/>
  <br/>
DELETE:  <br/>
in:  <br/>
    curl -X DELETE http://arch.homework/api/users/1  <br/>
out:  <br/>
    User [id=1] was deleted  <br/>
  <br/>
GET USER:  <br/>
in:  <br/>
    curl -X GET http://arch.homework/api/users/1  <br/>
out:  <br/>
    Failed to get user  <br/>
  <br/>
