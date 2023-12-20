let socket = new WebSocket("ws://127.0.0.1:8080/ws");
console.log("Attempting Connection...");
document.getElementById("input_form").addEventListener("submit", function (event) {
    event.preventDefault();
    //获取输入框的值
    var input_value = document.getElementById("input").value;
    message = {
        Type: 'text',
        Content: input_value
    }
    console.log(input_value)
    socket.send(JSON.stringify(message));
})

socket.onopen = () => {
    console.log("Successfully Connected");

};
socket.onclose = event => {
    console.log("Socket Closed Connection: ", event);
    socket.send("Client Closed!");
};
socket.onmessage = event => {
    console.log(event);
    let para = document.createElement("p");

// 解析 JSON 数据并将其赋值给 para.innerText
    let parsedData = JSON.parse(event.data);
    if (parsedData.type == "text") {
        para.innerText = parsedData.content; // 访问对象的 content 属性
        let tb = document.getElementById("data");
        tb.appendChild(para);
    }// 使用appendChild追加到末尾
    // 如果是点击切换私聊的话
    if (parsedData.type == "names") {
        let names = JSON.parse(parsedData.content);
        let membersContainer = document.getElementById("members");
        // 清空子节点
        membersContainer.innerHTML = '';
        let title = document.createElement("p");
        title.innerText = "在线成员列表(点击对应成员即可切换私聊)";
        membersContainer.appendChild(title);
        for (var i = 0; i < names.length ; i++) {
            let memberElement = document.createElement("p");
            memberElement.innerText = names[i];
            membersContainer.appendChild(memberElement);
        }

        console.log(names);
    }

    console.log(names);



}

socket.onerror = error => {
    console.log("Socket Error: ", error);
};
//向服务器获取在线用户人数及ID
