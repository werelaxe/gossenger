function addUser(id, nickname, first_name, last_name) {
    const usersContent = $("#users");
    const newRowElement = $(`<tr><td><a href="/user_page?user_id=${id}">${nickname}</a></td><td>${first_name}</td><td>${last_name}</td></tr>`);
    usersContent.append(newRowElement);
}


function setListUsersHandler(filter) {
    console.log("call it: " + filter);
    $.get("/users/search?filter=" + filter)
        .fail(function (data) {
            console.log("Fail while loading users");
            console.log(data)
        })
        .done(function (data) {
            const users = JSON.parse(data);
            const usersContent = $("#users");
            usersContent.empty();
            users.forEach(function (user) {
                addUser(user["id"], user["nickname"], user["first_name"], user["last_name"]);
            })
        });
}


function setSearchHandler() {
    const usersFilterFld = $("#users-filter-fld");
    usersFilterFld.on("keyup paste", function () {
        const searchFilter = usersFilterFld.val();
        if (searchFilter.length >= 3) {
            setListUsersHandler(searchFilter);
        }
    });
}


$(document).ready(function() {
    setSearchHandler();
});
