let ws = null;
let activeChatId = null;
let displayNames = {};


function waitSocket(socket, callback) {
    setTimeout(
        function () {
            let done = false;
            if (socket) {
                if (socket.readyState === 1) {
                    callback();
                    done = true;
                }
            }
            if (!done) {
                waitSocket(socket, callback);
            }
        },
        5);
}


function setMessagesReceivingHandler() {
    ws.onmessage = function (e) {
        const message = JSON.parse(e.data);
        if (activeChatId === message["chat_id"]) {
            addMessage(message["text"], message["sender_id"], message["time"]);
        }
    }
}


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


function saveNames(users) {
    users.forEach(function (user) {
        displayNames[user["id"]] = user["first_name"] + " " + user["last_name"];
    });
}


function getSeparatedDisplayNames(users) {
    return users.map(function (user) {
        return displayNames[user["id"]];
    }).join(", ");
}


function loadChatMembers(chatId) {
    jQuery.ajax({
        url: "/chats/list_members?chat_id=" + chatId,
        success: function (data) {
            const chat = getChat(chatId);
            const users = JSON.parse(data);
            saveNames(users);
            $("#chat-caption").text(chat.text() + " [" + getSeparatedDisplayNames(users) + "]");
        },
        error: function (data) {
            console.log("Fail while listing chat members");
            console.log(data)
        },
        async: false
    });
}


function activateChat(id) {
    loadChatMembers(id);
    clearMessages();
    loadMessages(id);
    showSender();

    let sendButton = $("#send-btn");
    sendButton.unbind("click");
    sendButton.on("click", function () {
        sendMessage(id, $("#message-inp").val());
    });

    activeChatId = id;
}


function sendMessage(chatId, text) {
    const sendMessageReq = JSON.stringify({
        "chat_id": chatId,
        "text": text
    });

    $.post("/messages/send", sendMessageReq)
        .fail(function (data) {
            console.log("Fail while sending a message");
            console.log(data)
        })
        .done(function (data) {
        });
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
    const newMessage = $("<li>" + displayNames[senderId] + ": '" + text + "', time: " + time + "</li>");
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


function clearMessages() {
    $("#messages").empty();
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
    hideSender();
    setIndexPageHandlers();
    loadChats();
    setMessagesReceivingHandler();
}


function hideSender() {
    $("#sender").hide();
}


function showSender() {
    $("#sender").show();
}


$(document).ready(function() {
    ws = new WebSocket("ws://" + location.host + "/messages_ws");
    waitSocket(ws, initIndexPage);
});
