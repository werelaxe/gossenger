let messagesWS = null;
let chatsWS = null;
let activeChatId = null;
let displayNames = new Map();
let chatTitles = {};
let pickedUsers = new Map();
let lastUsersSearch = null;


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


function getDisplayTime(unixTimestamp) {
    const date = new Date(unixTimestamp * 1000);
    const hours = date.getHours();
    const minutes = "0" + date.getMinutes();
    return hours + ':' + minutes.substr(-2);
}


function translateSymbols(text) {
    return text.replace("\n", "<br>");
}


function setMessagesReceivingHandler() {
    messagesWS.onmessage = function (e) {
        const message = JSON.parse(e.data);
        const senderId = message["sender_id"];
        const chatId = message["chat_id"];
        const messageText = message["text"];

        if (activeChatId === chatId) {
            addMessage(messageText, senderId, message["time"]);
        }

        const chatMessagePreview = $(`#chat-${chatId}-message-preview-text`);
        chatMessagePreview.text(`${getDisplayName(senderId)}: ${messageText}`);
        moveChatToTop(chatId);
    }
}


function setChatsReceivingHandler() {
    chatsWS.onmessage = function (e) {
        const chat = JSON.parse(e.data);
        addChat(chat["title"], chat["id"], chat["preview_message_text"], chat["preview_message_sender"]);
        moveChatToTop(chat["id"]);
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
        displayNames.set(user["id"], user["first_name"] + " " + user["last_name"]);
    });
}


function loadChatMembers(chatId) {
    $.ajax({
        url: "/chats/list_members?chat_id=" + chatId,
        success: function (data) {
            const users = JSON.parse(data);
            saveNames(users);
        },
        error: function (data) {
            console.log("Fail while listing chat members");
            console.log(data)
        },
        async: false
    });
}

function activateChat(id) {
    if (activeChatId === id) {
        return;
    }
    $("#chat-creating").hide();
    $("#prompt").hide();

    const messageInp = $("#message-inp");
    const sendButton = $("#send-btn");

    if (id === null) {
        getChat(activeChatId).removeClass("active-chat disable-hover");
        setMainContentTitle("Messages");
        hideSender();
        activeChatId = id;
        return;
    }
    loadChatMembers(id);
    clearMessages();
    loadMessages(id);
    showSender();

    sendButton.unbind("click");
    sendButton.on("click", function () {
        const messageText = messageInp.val();
        if (messageText.replace(/\s/g,'').length === 0) {
            return;
        }
        sendMessage(id, messageText);
        messageInp.val("");
    });

    messageInp.unbind("keydown");
    messageInp.on("keydown", function (event) {
        if (event.keyCode === 13) {
            if (event.ctrlKey || event.metaKey) {
                messageInp.val(messageInp.val() + "\n");
            } else {
                const messageText = messageInp.val();
                if (messageText.replace(/\s/g,'').length === 0) {
                    return;
                }
                sendMessage(id, messageText);
            }
        }
    });

    messageInp.unbind("keyup");
    messageInp.on("keyup", function (event) {
        if (event.keyCode === 13) {
            messageInp.val("");
        }
    });

    getChat(activeChatId).removeClass("active-chat disable-hover");
    getChat(id).addClass("active-chat disable-hover");
    setMainContentTitle("Messages of " + chatTitles[id]);
    activeChatId = id;
}


function setMainContentTitle(text) {
    $("#main-content-title").text(text);
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

function getDisplayName(userId) {
    if (!displayNames.has(userId)) {
        $.ajax({
            url: "/users/show?user_id=" + userId,
            success: function (data) {
                const user = JSON.parse(data);
                displayNames.set(user["id"], user["first_name"] + " " + user["last_name"]);
            },
            error: function (data) {
                console.log("Fail while getting user info");
                console.log(data)
            },
            async: false
        });
    }
    return displayNames.get(userId)
}

function addChat(title, id, previewMessageText, previewMessageSender) {
    const senderDisplayName = getDisplayName(previewMessageSender);

    if (previewMessageText === "") {
        addChatElement(title, id, `${senderDisplayName} created chat ${title}`);
    } else {
        addChatElement(title, id, `${senderDisplayName}: ${previewMessageText}`);
    }
}

function moveChatToTop(chatId) {
    const chat = getChat(chatId);
    chat.remove();
    $("#chats").prepend(chat);
}

function addChatElement(title, id, messagePreview) {
    const chatsDiv = $("#chats");
    const newChat = $(`
        <div class="chat-box" id="chat-${id}">
            <div class="chat-title">
                <span>${title}</span>
            </div>
            <div class="chat-message-preview">
                <span id="chat-${id}-message-preview-text">${messagePreview}</span>
            </div>
        </div>
    `);
    newChat.on("click", function () {
        activateChat(id);
    });
    chatsDiv.append(newChat);
    chatTitles[id] = title;
}


function addMessage(text, senderId, time) {
    const messagesDiv = $("#messages");
    const newMessage = $(`
        <div class="message-box">
            <div>
                <span class="message-sender-name"><a href="/user_page?user_id=${senderId}">${displayNames.get(senderId)}</a></span>
                <span class="message-time">${getDisplayTime(time)}</span>
            </div>
            <div class="message-text">
                <span>${translateSymbols(text)}</span>
            </div>
        </div>
    `);
    messagesDiv.append(newMessage);
    fullScrollMessages();
}


function loadChats() {
    $.get("/chats/list")
        .fail(function (data) {
            console.log("Fail while loading chats");
            console.log(data)
        })
        .done(function (data) {
            const chats = JSON.parse(data);
            if (chats === null) {
                console.log("Chats == null");
                return;
            }
            chats.forEach(function (chat) {
                addChat(chat["title"], chat["chat_id"], chat["preview_message_text"], chat["preview_message_sender"]);
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


function setLogoutButtonHandler() {
    let logoutButton = $("#logout-el");
    logoutButton.on("click", function () {
        clearCookies();
        location.reload();
    });
}


function setCreateChatHandler() {
    const createChatButton = $("#create-chat-btn");

    createChatButton.on("click", function () {
        const title = $("#new-chat-title-inp").val();
        const members = Array.from(pickedUsers.keys());
        const createChatReq = JSON.stringify({
            "title": title,
            "members": members
        });
        $.post("/chats/create", createChatReq)
            .fail(function (data) {
                console.log("Fail while creating a chat");
                console.log(data)
            })
            .done(function (data) {
                resetElements();
                const chatId =  JSON.parse(data)["chat_id"];
                activateChat(chatId);
            })
    });
}

function addFoundUser(user) {
    const displayName = user["first_name"] + " " + user["last_name"];
    displayNames.set(user["id"], displayName);
    const newButton = $(`<button type="submit" class="btn btn-default">Add</button>`);
    newButton.on("click", function () {
        pickedUsers.set(user["id"], user);
        updatePickedUsers();
        updateSearchUsers();
        $("#search-users-inp").val("")
    });
    addUserToTable($("#search-table"), displayName, user["nickname"], user["id"], newButton);
}


function addPickedUser(user) {
    const displayName = user["first_name"] + " " + user["last_name"];
    displayNames.set(user["id"], displayName);
    const newButton = $(`<button type="submit" class="btn btn-default">Remove</button>`);
    newButton.on("click", function () {
        pickedUsers.delete(user["id"]);
        updatePickedUsers();
        updateSearchUsers();
    });
    addUserToTable($("#picked-users"), displayName, user["nickname"], user["id"], newButton);
}


function addUserToTable(table, displayName, nickname, id, button) {
    const newTr = $(`<tr></tr>`);
    newTr.append($(`<td>${displayName}</td>`));
    newTr.append($(`<td>${nickname}</td>`));
    const newTd = $(`<td></td>`);
    newTd.append(button);
    newTr.append(newTd);
    table.append(newTr);
}


function clearFoundUsers() {
    $("#search-table").empty();
}

function clearPickedUsers() {
    $("#picked-users").empty();
}

function updateSearchUsers() {
    clearFoundUsers();
    lastUsersSearch.forEach(function (user) {
        if (!pickedUsers.has(user["id"])) {
            addFoundUser(user);
        }
    });
}

function updatePickedUsers() {
    clearPickedUsers();
    pickedUsers.forEach(function (user) {
        addPickedUser(user);
    });
}

function resetElements() {
    activateChat(null);
    $("#chat-creating").hide();
    clearMessages();
    $("#prompt").show();
    hideSender();
    setMainContentTitle("Messages");
}

function setShowChatCreatingContentHandler() {
    const handler = function () {
        $("#prompt").hide();
        clearMessages();
        activateChat(null);
        $("#chat-creating").show();
        setMainContentTitle("Select users for a new chat");

        const searchUsersInput = $("#search-users-inp");
        searchUsersInput.on("keyup paste", function () {
            const filter = searchUsersInput.val();
            if (filter.length < 3) {
                clearFoundUsers();
                return;
            }
            $.get("/users/search?filter=" + filter)
                .fail(function (data) {
                    console.log("Fail while searching users");
                    console.log(data)
                })
                .done(function (data) {
                    lastUsersSearch = JSON.parse(data);
                    updateSearchUsers();
                })
        });
    };
    $("#show-create-chat-btn-1").on("click", handler);
    $("#show-create-chat-btn-2").on("click", handler);
}


function setIndexPageHandlers() {
    setLogoutButtonHandler();
    setCreateChatHandler();
}


function normalizeMessagesHeight() {
    $("#messages").css("height", $(window).height() - 205);
}


function normalizeChatsHeight() {
    $("#chats").css("height", $(window).height() - 150);
}


function fullScrollMessages() {
    const messages = $("#messages");
    messages.scrollTop(messages[0].scrollHeight);
}


function setResizingHandlers() {
    $(window).resize(normalizeMessagesHeight);
    $(window).resize(normalizeChatsHeight);
}


function initIndexPage() {
    hideSender();
    setIndexPageHandlers();
    loadChats();
    setMessagesReceivingHandler();
    setChatsReceivingHandler();
    normalizeMessagesHeight();
    normalizeChatsHeight();
    setResizingHandlers();
    setShowChatCreatingContentHandler();
}


function hideSender() {
    $("#sender").hide();
}


function showSender() {
    $("#sender").show();
}


function waitChatSocket() {
    waitSocket(chatsWS, initIndexPage);
}


$(document).ready(function() {
    messagesWS = new WebSocket("ws://" + location.host + "/messages_ws");
    chatsWS = new WebSocket("ws://" + location.host + "/chats_ws");
    waitSocket(messagesWS, waitChatSocket);
});
