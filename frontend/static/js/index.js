function clearCookies() {
    const cookies = document.cookie.split(";");
    for (let i = 0; i < cookies.length; i++) {
        const cookie = cookies[i];
        const eqPos = cookie.indexOf("=");
        const name = eqPos > -1 ? cookie.substr(0, eqPos) : cookie;
        document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:00 GMT";
    }
}


function getChat(id) {
    return $("#chat-" + id);
}


function activateChat(id) {
    const chat = getChat(id);
    console.log("Make chat '" + chat.text() + "' active");
    $("#chat-caption").text(chat.text() + ":");
    loadMessages(id);
}


function addChat(name, id) {
    const chatsDiv = $("#chats");
    const newDiag = $("<li id='chat-" + id + "'>" + name + "</li>");
    newDiag.on("click", function () {
        activateChat(id);
    });
    chatsDiv.append(newDiag);
}


function addMessage(text, senderId, time) {
    const messagesDiv = $("#messages");
    const newMessage = $("<li>" + "User-" + senderId + ": '" + text + "', time: " + time + "</li>");
    messagesDiv.append(newMessage);
}

function loadChats() {
    $.get("/chats/list")
        .fail(function (data) {
            console.log("Fail while loading chats");
            console.log(data)
        })
        .done(function (data) {
            const chats = JSON.parse(data);
            chats.forEach(function (chat) {
                addChat(chat["title"], chat["chat_id"]);
            })
        });
}


function loadMessages(chatId) {
    $.get("/messages/list?chat_id=" + chatId)
        .fail(function (data) {
            console.log("Fail while loading messages");
            console.log(data)
        })
        .done(function (data) {
            const messages = JSON.parse(data);
            messages.forEach(function (message) {
                addMessage(message["text"], message["sender_id"], message["time"]);
            });
        });
}

function setIndexPageHandlers() {
    let logoutButton = $("#logout-btn");
    logoutButton.on("click", function () {
        clearCookies();
        location.reload();
    });
}


function initIndexPage() {
    setIndexPageHandlers();
    loadChats();
}

$(document).ready(function() {
    initIndexPage();
});
