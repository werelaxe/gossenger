let isLoginCorrect = false;
let isPasswordCorrect = false;
let isFirstNameCorrect = false;
let isLastNameCorrect = false;


function setRegisterButtonAbility() {
    $("#register-btn").prop("disabled", !(isLoginCorrect && isPasswordCorrect && isFirstNameCorrect && isLastNameCorrect));
}


function isNicknameBusy(nickname) {
    return  $.ajax({
        url: "/users/show?nickname=" + nickname,
        async: false
    }).status / 100 <= 3;
}


function setRegisterButtonHandlers() {
    let registerButton = $("#register-btn");
    registerButton.on("click", function () {
        let nicknameField = $("#nickname-inp");
        let nickname = nicknameField.val();
        let password = $("#password-inp").val();
        let firstName = $("#first-name-inp").val();
        let lastName = $("#last-name-inp").val();

        if (!(isLoginCorrect && isPasswordCorrect && isFirstNameCorrect && isLastNameCorrect)) {
            return false;
        }

        if (isNicknameBusy(nickname)) {
            showTooltip(nicknameField, "This nickname is busy");
            return false;
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


function setNicknameCheckingHandler() {
    let nicknameField = $("#nickname-inp");
    nicknameField.on("change", function () {
        if (isValid(nicknameField, nicknamePattern)) {
            if (isNicknameBusy(nicknameField.val())) {
                showTooltip(nicknameField, "This nickname is busy");
            }
        }
    });
}


function initRegisterPage() {
    setRegisterButtonHandlers();
    setRegisterValidatorHandlers();
    setNicknameCheckingHandler();
}


$(document).ready(function() {
    initRegisterPage();
});
