function addUser(id, nickname, first_name, last_name) {
    const usersDiv = $("#users");
    const newRowElement = $(`<tr><td>${id}</td><td>${nickname}</td><td>${first_name}</td><td>${last_name}</td></tr>`);
    usersDiv.append(newRowElement);
}


function setListUsersHandler() {
    $.get("/users/list")
        .fail(function (data) {
            console.log("Fail while loading users");
            console.log(data)
        })
        .done(function (data) {
            const users = JSON.parse(data);
            users.forEach(function (user) {
                addUser(user["id"], user["nickname"], user["first_name"], user["last_name"]);
            })
        });
}


$(document).ready(function() {
    setListUsersHandler();
});
