let isLoginCorrect = false;
let isPasswordCorrect = false;
let isFirstNameCorrect = false;
let isLastNameCorrect = false;


function setRegisterButtonAbility() {
    $("#register-btn").prop("disabled", !(isLoginCorrect && isPasswordCorrect && isFirstNameCorrect && isLastNameCorrect));
}


function setRegisterButtonHandlers() {
    let registerButton = $("#register-btn");
    registerButton.on("click", function () {
        let nickname = $("#nickname-inp").val();
        let password = $("#password-inp").val();
        let firstName = $("#first-name-inp").val();
        let lastName = $("#last-name-inp").val();

        if (!(isLoginCorrect && isPasswordCorrect && isFirstNameCorrect && isLastNameCorrect)) {
            return;
        }

        let registerReq = JSON.stringify({
            "nickname": nickname,
            "password": password,
            "first_name": firstName,
            "last_name": lastName
        });
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


function setRegisterValidatorHandlers() {
    setValidator(
        "nickname-inp",
        nicknamePattern,
        `Nickname must match regexp: ${nicknamePattern}`,
        function () {
            isLoginCorrect = false;
            setRegisterButtonAbility();
        },
        function () {
            isLoginCorrect = true;
            setRegisterButtonAbility();
        }
    )();
    setValidator(
        "password-inp",
        passwordPattern,
        `Password must match regexp: ${passwordPattern}`,
        function () {
            isPasswordCorrect = false;
            setRegisterButtonAbility();
        },
        function () {
            isPasswordCorrect = true;
            setRegisterButtonAbility();
        }
    )();
    setValidator(
        "first-name-inp",
        firstNamePattern,
        `First name must match regexp: ${firstNamePattern}`,
        function () {
            isFirstNameCorrect = false;
            setRegisterButtonAbility();
        },
        function () {
            isFirstNameCorrect = true;
            setRegisterButtonAbility();
        }
    )();
    setValidator(
        "last-name-inp",
        lastNamePattern,
        `Last name must match regexp: ${lastNamePattern}`,
        function () {
            isLastNameCorrect = false;
            setRegisterButtonAbility();
        },
        function () {
            isLastNameCorrect = true;
            setRegisterButtonAbility();
        }
    )();
}


function initRegisterPage() {
    setRegisterButtonHandlers();
    setRegisterValidatorHandlers();
}


$(document).ready(function() {
    initRegisterPage();
});
