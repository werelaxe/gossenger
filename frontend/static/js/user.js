function setSendPrivateMessageHandler() {
    $("#send-private-message-btn").on("click", function () {
        location.href = "/?ensure_private_chat=" + privateUserId.toString();
    });
}


$(document).ready(function() {
    setSendPrivateMessageHandler();
});
