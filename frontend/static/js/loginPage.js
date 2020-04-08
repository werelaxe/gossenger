function setLoginButtonHandler() {
    let loginButton = $("#login-btn");
    loginButton.on("click", function () {
        let nickname = $("#username-inp").val();
        let password = $("#password-inp").val();

        let loginReq = JSON.stringify({
            "nickname": nickname,
            "password": password
        });
        $.post("/login", loginReq)
            .fail(function (data) {
                console.log("Fail while logging");
                console.log(data)
            })
            .done(function (data) {
                document.location.href = "/";
            });
        return false;
    });
}


function setRegisterPageButtonHandler() {
    let registerPageButton = $("#register-page-btn");
    registerPageButton.on("click", function () {
        location.href = "/register_page";
        return false;
    });
}


function initLoginPage() {
    setLoginButtonHandler();
    setRegisterPageButtonHandler();
}

$(document).ready(function() {
    initLoginPage();
});
