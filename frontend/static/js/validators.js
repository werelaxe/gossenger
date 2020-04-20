const passwordPattern = /^[ -~]{3,20}$/;
const nicknamePattern = /^[0-9a-zA-Z]{3,20}$/;
const firstNamePattern = /^[a-zA-Z]{3,20}$/;
const lastNamePattern = /^[a-zA-Z]{3,20}$/;


function showTooltip(field, text) {
    field.tooltip({
        placement: 'right',
        trigger: 'manual',
        title: text,
    }).tooltip('show');
}


function hideTooltip(field) {
    field.tooltip('destroy');
}


function isValid(field, pattern) {
    return field.val().match(pattern);
}


function getValidator(fieldId, pattern, errorMessage, errorCallback, successCallback) {
    const field = $("#" + fieldId);
    return function () {
        if (!isValid(field, pattern)) {
            showTooltip(field, errorMessage);
            if (errorCallback !== undefined) {
                errorCallback();
            }
            return false;
        } else {
            hideTooltip(field);
            if (successCallback !== undefined) {
                successCallback();
            }
            return true;
        }
    };
}


function setValidator(fieldId, pattern, errorMessage, errorCallback, successCallback) {
    const validator = getValidator(fieldId, pattern, errorMessage, errorCallback, successCallback);
    $("#" + fieldId).on("change", validator);
    return validator;
}
