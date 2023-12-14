let socket = new WebSocket("ws://127.0.0.1:8080/ws");
console.log("Attempting Connection...");
document.getElementById("input_form").addEventListener("submit", function (event) {
    event.preventDefault();
    //获取输入框的值
    var input_value = document.getElementById("input").value;
    console.log(input_value)
    socket.send(input_value);
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
    if (parsedData.type == "text")
    {para.innerText = parsedData.content; // 访问对象的 content 属性
    let tb = document.getElementById("data");
    tb.appendChild(para); }// 使用appendChild追加到末尾
    if (parsedData.type == "numberlist")
    {console.log(parsedData.content);}




}

socket.onerror = error => {
    console.log("Socket Error: ", error);
};