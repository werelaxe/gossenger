function initLoginPage() {
    let loginButton = $("#lgn-btn");
    loginButton.on("click", function () {
        let nickname = $("#usnm").val();
        let password = $("#pswd").val();

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
