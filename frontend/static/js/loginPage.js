function initLoginPage() {
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

$(document).ready(function() {
    initLoginPage();
});
