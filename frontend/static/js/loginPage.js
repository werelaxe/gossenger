let isLoginCorrect = false;
let isPasswordCorrect = false;


function setLoginButtonAbility() {
    $("#login-btn").prop("disabled", !(isLoginCorrect && isPasswordCorrect));
}


function setLoginButtonHandler() {
    $("#login-btn").on("click", function () {
        let nicknameInp = $("#nickname-inp");
        let nickname = nicknameInp.val();
        let passwordInp = $("#password-inp");
        let password = passwordInp.val();

        if (!isLoginCorrect || !isPasswordCorrect) {
            return;
        }

        let loginReq = JSON.stringify({
            "nickname": nickname,
            "password": password
        });
        $.post("/login", loginReq)
            .fail(function (data) {
                console.log("Fail while logging");
                console.log(data);
                showTooltip(passwordInp, "Incorrect login or password");
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


function setValidatorHandlers() {
    setValidator(
        "nickname-inp",
        nicknamePattern,
        `Nickname must match regexp: ${nicknamePattern}`,
        function () {
            isLoginCorrect = false;
            setLoginButtonAbility();
        },
        function () {
            isLoginCorrect = true;
            setLoginButtonAbility();
        }
    )();
    setValidator(
        "password-inp",
        passwordPattern,
        `Password must match regexp: ${passwordPattern}`,
        function () {
            isPasswordCorrect = false;
            setLoginButtonAbility();
        },
        function () {
            isPasswordCorrect = true;
            setLoginButtonAbility();
        }
    )();
}


function initLoginPage() {
    setLoginButtonHandler();
    setRegisterPageButtonHandler();
    setValidatorHandlers();
}

$(document).ready(function() {
    initLoginPage();
});
