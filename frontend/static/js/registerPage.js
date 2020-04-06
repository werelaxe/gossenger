function initRegisterPage() {
    let registerButton = $("#register-btn");
    registerButton.on("click", function () {
        let nickname = $("#username-inp").val();
        let password = $("#password-inp").val();
        let firstName = $("#first-name-inp").val();
        let lastName = $("#last-name-inp").val();

        let registerReq = JSON.stringify({
            "nickname": nickname,
            "password": password,
            "first_name": firstName,
            "last_name": lastName
        });
        console.log(registerReq);
        $.post("/register", registerReq)
            .fail(function (data) {
                console.log("Fail while registration");
                console.log(data)
            })
            .done(function (data) {
                document.location.href = "/";
            });
        return false;
    });
}

$(document).ready(function() {
    initRegisterPage();
});
